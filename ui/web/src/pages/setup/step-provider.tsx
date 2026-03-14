import { useState, useMemo } from "react";
import { useTranslation } from "react-i18next";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Card, CardContent } from "@/components/ui/card";
import { TooltipProvider } from "@/components/ui/tooltip";
import { InfoTip } from "@/pages/setup/info-tip";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { PROVIDER_TYPES } from "@/constants/providers";
import { useProviders } from "@/pages/providers/hooks/use-providers";
import { CLISection } from "@/pages/providers/provider-cli-section";
import { slugify } from "@/lib/slug";
import type { ProviderData } from "@/types/provider";

interface StepProviderProps {
  providers?: ProviderData[];
  onComplete: (provider: ProviderData) => void;
  onBack?: () => void;
}

export function StepProvider({ providers = [], onComplete, onBack }: StepProviderProps) {
  const { t } = useTranslation("setup");
  const { createProvider, updateProvider } = useProviders();

  const [providerType, setProviderType] = useState("openrouter");
  const [displayName, setDisplayName] = useState("openrouter");
  const [apiKey, setApiKey] = useState("");
  const [apiBase, setApiBase] = useState("https://openrouter.ai/api/v1");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [duplicateMode, setDuplicateMode] = useState<"create" | "use" | "update" | null>(null);
  // Track if user chose to create new (ignore duplicate)
  const [ignoreDuplicate, setIgnoreDuplicate] = useState(false);

  // Slugified name for API
  const name = useMemo(() => slugify(displayName), [displayName]);

  // Find existing provider with same name (skip if user chose to create new)
  const existingProvider = useMemo(
    () => ignoreDuplicate ? null : providers.find((p) => slugify(p.name) === name && name.trim() !== ""),
    [providers, name, ignoreDuplicate],
  );

  // When user manually changes displayName, reset ignoreDuplicate to re-check
  const handleDisplayNameChange = (value: string) => {
    setDisplayName(value);
    setIgnoreDuplicate(false);
    setError("");
  };

  const isCLI = providerType === "claude_cli";
  // Local Ollama uses no API key — the server accepts any non-empty Bearer value internally
  const isOllama = providerType === "ollama";

  const handleTypeChange = (value: string) => {
    setProviderType(value);
    const preset = PROVIDER_TYPES.find((t) => t.value === value);
    setDisplayName(value);
    setApiBase(preset?.apiBase || "");
    setApiKey("");
    setError("");
    setDuplicateMode(null);
    setIgnoreDuplicate(false);
  };

  const apiBasePlaceholder = useMemo(
    () => PROVIDER_TYPES.find((t) => t.value === providerType)?.placeholder
      || PROVIDER_TYPES.find((t) => t.value === providerType)?.apiBase
      || "https://api.example.com/v1",
    [providerType],
  );

  const handleCreate = async () => {
    if (!isCLI && !isOllama && !apiKey.trim()) { setError(t("provider.errors.apiKeyRequired")); return; }

    // Check for duplicate
    if (existingProvider && !ignoreDuplicate) {
      if (duplicateMode === "use") {
        onComplete(existingProvider);
        return;
      }
      if (duplicateMode === "update") {
        setLoading(true);
        setError("");
        try {
          const updated = await updateProvider(existingProvider.id, {
            provider_type: providerType,
            api_base: apiBase.trim() || undefined,
            api_key: isCLI || isOllama ? undefined : apiKey.trim(),
            enabled: true,
          });
          onComplete(updated);
        } catch (err) {
          setError(err instanceof Error ? err.message : t("provider.errors.failedCreate"));
        } finally {
          setLoading(false);
        }
        return;
      }
      if (duplicateMode === "create") {
        // User chose to create new anyway - proceed
      } else {
        // First time detecting duplicate - show options to user
        setDuplicateMode("create");
        setError(t("provider.errors.duplicateName"));
        return;
      }
    }

    setLoading(true);
    setError("");
    try {
      const provider = await createProvider({
        name: name.trim(),
        provider_type: providerType,
        api_base: apiBase.trim() || undefined,
        api_key: isCLI || isOllama ? undefined : apiKey.trim(),
        enabled: true,
      }) as ProviderData;
      onComplete(provider);
    } catch (err) {
      // Check if it's a duplicate error
      const errMsg = err instanceof Error ? err.message : "";
      if (errMsg.includes("duplicate key") || errMsg.includes("unique constraint")) {
        setDuplicateMode("create");
        setError(t("provider.errors.duplicateName"));
      } else {
        setError(errMsg || t("provider.errors.failedCreate"));
      }
    } finally {
      setLoading(false);
    }
  };

  const handleUseExisting = () => {
    if (existingProvider) {
      onComplete(existingProvider);
    }
  };

  const handleUpdateExisting = () => {
    setDuplicateMode("update");
    setError("");
  };

  return (
    <Card>
      <CardContent className="space-y-4 pt-6">
        <TooltipProvider>
          <div className="space-y-1">
            <h2 className="text-lg font-semibold">{t("provider.title")}</h2>
            <p className="text-sm text-muted-foreground">
              {isCLI
                ? t("provider.descriptionCli")
                : t("provider.description")}
            </p>
          </div>

          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <div className="space-y-2">
              <Label className="inline-flex items-center gap-1.5">
                {t("provider.providerType")}
                <InfoTip text={t("provider.providerTypeHint")} />
              </Label>
              <Select value={providerType} onValueChange={handleTypeChange}>
                <SelectTrigger><SelectValue /></SelectTrigger>
                <SelectContent>
                  {PROVIDER_TYPES.map((t) => (
                    <SelectItem key={t.value} value={t.value}>{t.label}</SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            <div className="space-y-2">
              <Label className="inline-flex items-center gap-1.5">
                {t("provider.name")}
                <InfoTip text={t("provider.nameHint")} />
              </Label>
              <Input value={displayName} onChange={(e) => handleDisplayNameChange(e.target.value)} />
            </div>
          </div>

          {isCLI ? (
            <CLISection open={true} />
          ) : (
            <>
              <div className="space-y-2">
                <Label className="inline-flex items-center gap-1.5">
                  {t("provider.apiKey")}
                  <InfoTip text={t("provider.apiKeyHint")} />
                </Label>
                <Input
                  type="password"
                  value={apiKey}
                  onChange={(e) => setApiKey(e.target.value)}
                  placeholder="sk-..."
                />
              </div>

              <div className="space-y-2">
                <Label className="inline-flex items-center gap-1.5">
                  {t("provider.apiBase")}
                  <InfoTip text={t("provider.apiBaseHint")} />
                </Label>
                <Input
                  value={apiBase}
                  onChange={(e) => setApiBase(e.target.value)}
                  placeholder={apiBasePlaceholder}
                />
              </div>
            </>
          )}

          {error && <p className="text-sm text-destructive">{error}</p>}

          {/* Duplicate provider options */}
          {duplicateMode && existingProvider && (
            <div className="rounded-md border border-amber-200 bg-amber-50 p-3 dark:border-amber-800 dark:bg-amber-950">
              <p className="mb-2 text-sm font-medium text-amber-700 dark:text-amber-300">
                {t("provider.duplicateFound", { name: existingProvider.name })}
              </p>
              <div className="flex flex-wrap gap-2">
                <Button size="sm" variant="outline" onClick={handleUseExisting}>
                  {t("provider.useExisting")}
                </Button>
                <Button size="sm" variant="outline" onClick={handleUpdateExisting}>
                  {t("provider.updateExisting")}
                </Button>
                <Button size="sm" variant="outline" onClick={() => { setDuplicateMode(null); setIgnoreDuplicate(true); setError(""); }}>
                  {t("provider.createNew")}
                </Button>
              </div>
            </div>
          )}

          <div className="flex justify-between">
            {onBack && (
              <Button variant="outline" onClick={onBack}>
                {t("common.back")}
              </Button>
            )}
            {!duplicateMode && (
              <Button onClick={handleCreate} disabled={loading || (!isCLI && !isOllama && !apiKey.trim())} className={onBack ? "" : "ml-auto"}>
                {loading ? t("provider.creating") : t("provider.create")}
              </Button>
            )}
          </div>
        </TooltipProvider>
      </CardContent>
    </Card>
  );
}
