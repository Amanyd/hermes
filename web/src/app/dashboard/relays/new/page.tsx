"use client";

import { useRouter } from "next/navigation";
import { useId, useState, useEffect } from "react";
import { toast } from "sonner";
import { useConnections, useCreateRelay, useSecrets } from "@/lib/queries";
import {
  ACTION_LABELS,
  ACTION_TYPES,
  type ActionType,
  type CreateRelayActionInput,
} from "@/types/relay";

const TEMPLATE_HINT =
  "Supports {{ payload.x }}, {{ prev.output.x }}, {{ steps[0].output.x }}";

function ConfigField({
  cfgKey,
  label,
  placeholder,
  hint,
  value,
  onChange,
  showTemplateHint = true,
}: {
  cfgKey: string;
  label: string;
  placeholder: string;
  hint?: string;
  value: string;
  onChange: (key: string, value: unknown) => void;
  showTemplateHint?: boolean;
}) {
  const id = useId();
  return (
    <div>
      <label
        htmlFor={id}
        className="mb-1 block text-xs font-medium text-zinc-400"
      >
        {label}
      </label>
      <input
        id={id}
        type="text"
        value={value}
        onChange={(e) => onChange(cfgKey, e.target.value)}
        placeholder={placeholder}
        className="w-full rounded-lg border border-white/10 bg-white/5 px-3 py-2 text-xs text-white placeholder:text-zinc-600 focus:border-orange-500/50 focus:outline-none focus:ring-1 focus:ring-orange-500/50"
      />
      {hint && <p className="mt-0.5 text-xs text-zinc-600">{hint}</p>}
      {showTemplateHint && (
        <p className="mt-0.5 text-xs text-zinc-700">{TEMPLATE_HINT}</p>
      )}
    </div>
  );
}

function SegmentedToggle({
  label,
  value,
  onChange,
  options,
}: {
  label: string;
  value: string;
  onChange: (value: string) => void;
  options: { label: string; value: string }[];
}) {
  return (
    <div>
      <p className="mb-2 block text-xs font-medium text-zinc-400">{label}</p>
      <div className="inline-flex rounded-lg border border-white/10 bg-[#111111] p-1">
        {options.map((option) => {
          const active = value === option.value;
          return (
            <button
              key={option.value}
              type="button"
              onClick={() => onChange(option.value)}
              className={`rounded-md px-3 py-1.5 text-xs font-medium transition-colors ${
                active
                  ? "bg-orange-500 text-white"
                  : "text-zinc-400 hover:text-white"
              }`}
            >
              {option.label}
            </button>
          );
        })}
      </div>
    </div>
  );
}

function ConfigSelectField({
  cfgKey,
  label,
  hint,
  value,
  options,
  placeholder = "Select an option",
  onChange,
}: {
  cfgKey: string;
  label: string;
  hint?: string;
  value: string;
  options: { label: string; value: string }[];
  placeholder?: string;
  onChange: (key: string, value: unknown) => void;
}) {
  const id = useId();

  return (
    <div>
      <label
        htmlFor={id}
        className="mb-1 block text-xs font-medium text-zinc-400"
      >
        {label}
      </label>
      <select
        id={id}
        value={value}
        onChange={(e) => onChange(cfgKey, e.target.value)}
        className="w-full rounded-lg border border-white/10 bg-[#1a1a1a] px-3 py-2 text-xs text-white focus:border-orange-500/50 focus:outline-none"
      >
        <option value="">{placeholder}</option>
        {options.map((option) => (
          <option key={option.value} value={option.value}>
            {option.label}
          </option>
        ))}
      </select>
      {hint && <p className="mt-0.5 text-xs text-zinc-600">{hint}</p>}
    </div>
  );
}

function DiscordConfigFields({
  config,
  onChange,
  secretOptions,
}: {
  config: Record<string, unknown>;
  onChange: (key: string, value: unknown) => void;
  secretOptions: { label: string; value: string }[];
}) {
  const [webhookMode, setWebhookMode] = useState(
    (config.webhook_url_ref as string)?.trim() !== "" ? "secret" : "direct",
  );

  useEffect(() => {
    setWebhookMode(
      (config.webhook_url_ref as string)?.trim() !== "" ? "secret" : "direct",
    );
  }, [config.webhook_url_ref]);

  const switchMode = (mode: string) => {
    setWebhookMode(mode);
    if (mode === "direct") {
      onChange("webhook_url_ref", "");
    } else {
      onChange("webhook_url", "");
    }
  };

  return (
    <div className="space-y-3">
      <SegmentedToggle
        label="Webhook Source"
        value={webhookMode}
        onChange={switchMode}
        options={[
          { label: "Direct URL", value: "direct" },
          { label: "Saved Secret", value: "secret" },
        ]}
      />

      {webhookMode === "direct" ? (
        <ConfigField
          cfgKey="webhook_url"
          label="Webhook URL"
          placeholder="https://discord.com/api/webhooks/..."
          value={(config.webhook_url as string) ?? ""}
          onChange={onChange}
          showTemplateHint={false}
        />
      ) : (
        <ConfigSelectField
          cfgKey="webhook_url_ref"
          label="Webhook URL Secret"
          placeholder="Select a saved secret"
          hint="Choose one of your saved secrets"
          value={(config.webhook_url_ref as string) ?? ""}
          options={secretOptions}
          onChange={onChange}
        />
      )}

      <ConfigField
        cfgKey="message_template"
        label="Message"
        placeholder="Hello {{payload.user.name}}"
        hint="Supports {{ payload.x }}, {{ prev.output.x }}, {{ steps[0].output.x }}"
        value={(config.message_template as string) ?? ""}
        onChange={onChange}
      />
    </div>
  );
}

function SlackConfigFields({
  config,
  onChange,
  secretOptions,
}: {
  config: Record<string, unknown>;
  onChange: (key: string, value: unknown) => void;
  secretOptions: { label: string; value: string }[];
}) {
  const [webhookMode, setWebhookMode] = useState(
    (config.webhook_url_ref as string)?.trim() !== "" ? "secret" : "direct",
  );

  useEffect(() => {
    setWebhookMode(
      (config.webhook_url_ref as string)?.trim() !== "" ? "secret" : "direct",
    );
  }, [config.webhook_url_ref]);

  const switchMode = (mode: string) => {
    setWebhookMode(mode);
    if (mode === "direct") {
      onChange("webhook_url_ref", "");
    } else {
      onChange("webhook_url", "");
    }
  };

  return (
    <div className="space-y-3">
      <SegmentedToggle
        label="Webhook Source"
        value={webhookMode}
        onChange={switchMode}
        options={[
          { label: "Direct URL", value: "direct" },
          { label: "Saved Secret", value: "secret" },
        ]}
      />

      {webhookMode === "direct" ? (
        <ConfigField
          cfgKey="webhook_url"
          label="Webhook URL"
          placeholder="https://hooks.slack.com/services/..."
          value={(config.webhook_url as string) ?? ""}
          onChange={onChange}
          showTemplateHint={false}
        />
      ) : (
        <ConfigSelectField
          cfgKey="webhook_url_ref"
          label="Webhook URL Secret"
          placeholder="Select a saved secret"
          hint="Choose one of your saved secrets"
          value={(config.webhook_url_ref as string) ?? ""}
          options={secretOptions}
          onChange={onChange}
        />
      )}

      <ConfigField
        cfgKey="message_template"
        label="Message"
        placeholder="Hello {{payload.user.name}}"
        hint="Supports {{ payload.x }}, {{ prev.output.x }}, {{ steps[0].output.x }}"
        value={(config.message_template as string) ?? ""}
        onChange={onChange}
      />
    </div>
  );
}

function ActionConfigFields({
  type,
  config,
  onChange,
}: {
  type: ActionType;
  config: Record<string, unknown>;
  onChange: (key: string, value: unknown) => void;
}) {
  const methodId = useId();
  const connectionId = useId();
  const { data: connections } = useConnections();
  const { data: secrets } = useSecrets();

  const secretOptions = (secrets ?? []).map((secret) => ({
    label: secret.name,
    value: secret.name,
  }));
  switch (type) {
    case "discord_send":
      return (
        <DiscordConfigFields
          config={config}
          onChange={onChange}
          secretOptions={secretOptions}
        />
      );

    case "slack_send":
      return (
        <SlackConfigFields
          config={config}
          onChange={onChange}
          secretOptions={secretOptions}
        />
      );

    case "http_request":
      return (
        <div className="space-y-2">
          <ConfigField
            cfgKey="url"
            label="URL"
            placeholder="https://example.com/endpoint"
            value={(config.url as string) ?? ""}
            onChange={onChange}
          />
          <div>
            <label
              htmlFor={methodId}
              className="mb-1 block text-xs font-medium text-zinc-400"
            >
              Method
            </label>
            <select
              id={methodId}
              value={(config.method as string) ?? "POST"}
              onChange={(e) => onChange("method", e.target.value)}
              className="w-full rounded-lg border border-white/10 bg-[#1a1a1a] px-3 py-2 text-xs text-white focus:border-orange-500/50 focus:outline-none"
            >
              {["GET", "POST", "PUT", "PATCH", "DELETE"].map((m) => (
                <option key={m} value={m}>
                  {m}
                </option>
              ))}
            </select>
          </div>
          <ConfigField
            cfgKey="body"
            label="Body (JSON)"
            placeholder='{"key": "value"}'
            value={(config.body as string) ?? ""}
            onChange={onChange}
          />
        </div>
      );
    case "email_send":
      return (
        <div className="space-y-2">
          <div>
            <label
              htmlFor={connectionId}
              className="mb-1 block text-xs font-medium text-zinc-400"
            >
              Email Connection <span className="text-orange-500">*</span>
            </label>
            <select
              id={connectionId}
              value={(config.connection_id as string) ?? ""}
              onChange={(e) => onChange("connection_id", e.target.value)}
              className="w-full rounded-lg border border-white/10 bg-[#1a1a1a] px-3 py-2 text-xs text-white focus:border-orange-500/50 focus:outline-none"
            >
              <option value="">Select a connected account…</option>
              {(connections ?? []).map((c) => (
                <option key={c.id} value={c.id}>
                  {c.provider === "google" ? "Google" : "Microsoft"}:{" "}
                  {c.account_email}
                </option>
              ))}
            </select>
            {(!connections || connections.length === 0) && (
              <p className="mt-0.5 text-xs text-amber-500/80">
                No connections yet — add one in the{" "}
                <a
                  href="/dashboard/connections"
                  className="underline hover:text-amber-400"
                >
                  Connections
                </a>{" "}
                page.
              </p>
            )}
          </div>
          <ConfigField
            cfgKey="to"
            label="To"
            placeholder="user@example.com or {{ payload.email }}"
            value={(config.to as string) ?? ""}
            onChange={onChange}
          />
          <ConfigField
            cfgKey="subject"
            label="Subject"
            placeholder="Hello from Hermes"
            value={(config.subject as string) ?? ""}
            onChange={onChange}
          />
          <ConfigField
            cfgKey="body"
            label="Body"
            placeholder="Hi {{ payload.name }}, your event fired!"
            value={(config.body as string) ?? ""}
            onChange={onChange}
          />
        </div>
      );
    case "debug_log":
      return (
        <p className="text-xs text-zinc-500">
          Logs the raw payload — no config needed.
        </p>
      );
    default:
      return null;
  }
}

function ActionCard({
  index,
  action,
  onUpdate,
  onRemove,
  canRemove,
}: {
  index: number;
  action: CreateRelayActionInput;
  onUpdate: (index: number, updated: CreateRelayActionInput) => void;
  onRemove: (index: number) => void;
  canRemove: boolean;
}) {
  const setType = (type: ActionType) =>
    onUpdate(index, { ...action, action_type: type, config: {} });

  const setConfig = (key: string, value: unknown) =>
    onUpdate(index, { ...action, config: { ...action.config, [key]: value } });

  return (
    <div className="rounded-xl border border-white/10 bg-[#141414] p-4">
      <div className="flex items-center justify-between mb-3">
        <span className="text-xs font-medium text-zinc-400">
          Action {index + 1}
        </span>
        {canRemove && (
          <button
            type="button"
            onClick={() => onRemove(index)}
            className="text-xs text-red-400 hover:text-red-300 transition-colors"
          >
            Remove
          </button>
        )}
      </div>
      <div className="mb-3">
        <label
          htmlFor={`action-type-${index}`}
          className="mb-1 block text-xs font-medium text-zinc-400"
        >
          Type
        </label>
        <select
          id={`action-type-${index}`}
          value={action.action_type}
          onChange={(e) => setType(e.target.value as ActionType)}
          className="w-full rounded-lg border border-white/10 bg-[#1a1a1a] px-3 py-2 text-sm text-white focus:border-orange-500/50 focus:outline-none"
        >
          {ACTION_TYPES.map((t) => (
            <option key={t} value={t}>
              {ACTION_LABELS[t]}
            </option>
          ))}
        </select>
      </div>
      <ActionConfigFields
        type={action.action_type as ActionType}
        config={action.config}
        onChange={setConfig}
      />
    </div>
  );
}

export default function NewRelayPage() {
  const router = useRouter();
  const createMutation = useCreateRelay();
  const nameId = useId();
  const descId = useId();
  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [actions, setActions] = useState<CreateRelayActionInput[]>([
    { action_type: "debug_log", config: {}, order_index: 0 },
  ]);

  const addAction = () =>
    setActions((prev) => [
      ...prev,
      { action_type: "debug_log", config: {}, order_index: prev.length },
    ]);

  const updateAction = (index: number, updated: CreateRelayActionInput) =>
    setActions((prev) => prev.map((a, i) => (i === index ? updated : a)));

  const removeAction = (index: number) =>
    setActions((prev) =>
      prev
        .filter((_, i) => i !== index)
        .map((a, i) => ({ ...a, order_index: i })),
    );

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!name.trim()) return toast.error("Name is required");
    try {
      const relay = await createMutation.mutateAsync({
        name,
        description,
        actions,
      });
      toast.success("Relay created!");
      router.push(`/dashboard/relays/${relay.id}`);
    } catch (err) {
      toast.error(
        err instanceof Error ? err.message : "Failed to create relay",
      );
    }
  };

  return (
    <div className="p-8 max-w-2xl">
      <div className="mb-8">
        <h1 className="text-xl font-bold text-white">New Relay</h1>
        <p className="mt-0.5 text-sm text-zinc-500">
          Configure a webhook endpoint and its actions
        </p>
      </div>

      <form onSubmit={handleSubmit} className="space-y-6">
        {/* Basic info */}
        <div className="rounded-xl border border-white/10 bg-[#1a1a1a] p-5 space-y-4">
          <h2 className="text-sm font-medium text-zinc-300">Basic info</h2>
          <div>
            <label
              htmlFor={nameId}
              className="mb-1.5 block text-sm font-medium text-zinc-300"
            >
              Name <span className="text-orange-500">*</span>
            </label>
            <input
              id={nameId}
              type="text"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="My webhook relay"
              className="w-full rounded-lg border border-white/10 bg-white/5 px-4 py-2.5 text-sm text-white placeholder:text-zinc-500 focus:border-orange-500/50 focus:outline-none focus:ring-1 focus:ring-orange-500/50"
            />
          </div>
          <div>
            <label
              htmlFor={descId}
              className="mb-1.5 block text-sm font-medium text-zinc-300"
            >
              Description
            </label>
            <input
              id={descId}
              type="text"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="Optional description"
              className="w-full rounded-lg border border-white/10 bg-white/5 px-4 py-2.5 text-sm text-white placeholder:text-zinc-500 focus:border-orange-500/50 focus:outline-none focus:ring-1 focus:ring-orange-500/50"
            />
          </div>
        </div>

        {/* Actions */}
        <div className="space-y-3">
          <div className="flex items-center justify-between">
            <h2 className="text-sm font-medium text-zinc-300">Actions</h2>
            <button
              type="button"
              onClick={addAction}
              className="flex items-center gap-1.5 text-xs text-orange-400 hover:text-orange-300 transition-colors"
            >
              <svg
                aria-hidden="true"
                className="h-3.5 w-3.5"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
                strokeWidth={2}
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  d="M12 4.5v15m7.5-7.5h-15"
                />
              </svg>
              Add action
            </button>
          </div>
          {actions.map((action, i) => (
            <ActionCard
              // biome-ignore lint/suspicious/noArrayIndexKey: order is stable here
              key={i}
              index={i}
              action={action}
              onUpdate={updateAction}
              onRemove={removeAction}
              canRemove={actions.length > 1}
            />
          ))}
        </div>

        {/* Submit */}
        <div className="flex items-center gap-3 pt-2">
          <button
            type="submit"
            disabled={createMutation.isPending}
            className="rounded-lg bg-orange-500 px-5 py-2.5 text-sm font-semibold text-white transition-colors hover:bg-orange-600 disabled:cursor-not-allowed disabled:opacity-50"
          >
            {createMutation.isPending ? "Creating…" : "Create relay"}
          </button>
          <button
            type="button"
            onClick={() => router.back()}
            className="rounded-lg border border-white/10 px-5 py-2.5 text-sm text-zinc-400 transition-colors hover:border-white/20 hover:text-white"
          >
            Cancel
          </button>
        </div>
      </form>
    </div>
  );
}
