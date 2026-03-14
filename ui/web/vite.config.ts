import { defineConfig, loadEnv } from "vite";
import react from "@vitejs/plugin-react";
import tailwindcss from "@tailwindcss/vite";
import path from "path";

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), "");
  const backendHost = env.VITE_BACKEND_HOST || "192.168.3.97";
  const backendPort = env.VITE_BACKEND_PORT || "18790";

  return {
    plugins: [react(), tailwindcss()],
    resolve: {
      alias: {
        "@": path.resolve(__dirname, "./src"),
      },
    },
    server: {
      host: "192.168.3.97",
      port: 5173,
      proxy: {
        "/ws": {
          target: `http://${backendHost}:${backendPort}`,
          ws: true,
          changeOrigin: true,
        },
        "/v1": {
          target: `http://${backendHost}:${backendPort}`,
          changeOrigin: true,
        },
        "/health": {
          target: `http://${backendHost}:${backendPort}`,
          changeOrigin: true,
        },
      },
    },
    build: {
      outDir: "dist",
      emptyOutDir: true,
    },
  };
});
