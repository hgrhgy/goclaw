import { useEffect, useRef, useMemo, useCallback } from "react";
import { WsClient, type ConnectionState } from "@/api/ws-client";
import { HttpClient } from "@/api/http-client";
import { WsContext } from "@/hooks/use-ws";
import { useAuthStore } from "@/stores/use-auth-store";
import { useWsQueryInvalidation } from "@/hooks/use-query-invalidation";
import { useWsEvent } from "@/hooks/use-ws-event";
import { TEAM_RELATED_EVENTS } from "@/api/protocol";
import { useTeamEventStore } from "@/stores/use-team-event-store";

// Use VITE_BACKEND_HOST and VITE_BACKEND_PORT to build absolute URLs for direct backend connection.
// If not set, use relative paths (goes through Vite proxy in dev, same-origin in prod).
// VITE_BACKEND_HOST: Backend IP/hostname (e.g., "192.168.3.97")
// VITE_BACKEND_PORT: Backend port (e.g., "18790")
// VITE_WS_URL: WebSocket URL - can be absolute (ws://host:port/path) or relative (/ws)
function getApiUrl(): string {
  const host = import.meta.env.VITE_BACKEND_HOST;
  const port = import.meta.env.VITE_BACKEND_PORT;
  if (host && port) {
    return `http://${host}:${port}`;
  }
  return ""; // Use relative path (Vite proxy in dev, same-origin in prod)
}

function getWsUrl(): string {
  const envWsUrl = import.meta.env.VITE_WS_URL;
  if (envWsUrl) {
    return envWsUrl;
  }
  const apiUrl = getApiUrl();
  if (apiUrl) {
    // Derive WebSocket URL from API URL
    const proto = apiUrl.startsWith("https") ? "wss" : "ws";
    const hostPort = apiUrl.replace(/^https?:\/\//, "");
    return `${proto}://${hostPort}/ws`;
  }
  return "/ws"; // Use relative path (Vite proxy)
}

const API_URL = getApiUrl();
const WS_URL = getWsUrl();

export function WsProvider({ children }: { children: React.ReactNode }) {
  const token = useAuthStore((s) => s.token);
  const userId = useAuthStore((s) => s.userId);
  const senderID = useAuthStore((s) => s.senderID);

  const wsRef = useRef<WsClient | null>(null);

  // Create WsClient once - survives StrictMode remounts
  if (!wsRef.current) {
    wsRef.current = new WsClient(
      WS_URL,
      () => useAuthStore.getState().token,
      () => useAuthStore.getState().userId,
      () => useAuthStore.getState().senderID,
      (state: ConnectionState) => {
        useAuthStore.getState().setConnected(state === "connected");
      },
    );
    wsRef.current.onAuthFailure = () => {
      useAuthStore.getState().logout();
    };
  }
  const ws = wsRef.current;

  const http = useMemo(() => {
    const client = new HttpClient(
      API_URL,
      () => useAuthStore.getState().token,
      () => useAuthStore.getState().userId,
    );
    client.onAuthFailure = () => {
      useAuthStore.getState().logout();
    };
    return client;
  }, []);

  // Auto-connect when credentials are available (token or sender_id), disconnect when not.
  useEffect(() => {
    if ((token || senderID) && userId) {
      ws.connect();
    } else {
      ws.disconnect();
    }
  }, [token, userId, senderID, ws]);

  const value = useMemo(() => ({ ws, http }), [ws, http]);

  return (
    <WsContext.Provider value={value}>
      <WsQueryInvalidation />
      <WsTeamEventCapture />
      {children}
    </WsContext.Provider>
  );
}

function WsQueryInvalidation() {
  useWsQueryInvalidation();
  return null;
}

/** Captures all team-related WS events into the Zustand store. */
function WsTeamEventCapture() {
  const addEvent = useTeamEventStore((s) => s.addEvent);

  const handler = useCallback(
    (raw: unknown) => {
      const { event, payload } = raw as { event: string; payload: unknown };
      if (!TEAM_RELATED_EVENTS.has(event)) return;
      // Skip noisy chunk/thinking subtypes for agent events
      if (event === "agent") {
        const p = payload as { type?: string };
        if (p.type === "chunk" || p.type === "thinking") return;
      }
      addEvent(event, payload);
    },
    [addEvent],
  );

  useWsEvent("*", handler);
  return null;
}
