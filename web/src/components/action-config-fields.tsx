"use client";

import { useId, useState, useEffect } from "react";
import { useConnections, useSecrets } from "@/lib/queries";
import {
  ACTION_LABELS,
  ACTION_TYPES,
  type ActionType,
  type CreateRelayActionInput,
} from "@/types/relay";

const TEMPLATE_HINT =
  "Supports {{ payload.x }}, {{ prev.output.x }}, {{ steps[0].output.x }}";

export function ConfigField({
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

export function SegmentedToggle({
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
    (config.webhook_url_ref as string)?.trim() ? "secret" : "direct",
  );

  useEffect(() => {
    setWebhookMode(
      (config.webhook_url_ref as string)?.trim() ? "secret" : "direct",
    );
  }, [config.webhook_url_ref]);

  const switchMode = (mode: string) => {
    setWebhookMode(mode);
    if (mode === "direct") onChange("webhook_url_ref", undefined);
    else onChange("webhook_url", undefined);
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
        showTemplateHint={false}
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
    (config.webhook_url_ref as string)?.trim() ? "secret" : "direct",
  );

  useEffect(() => {
    setWebhookMode(
      (config.webhook_url_ref as string)?.trim() ? "secret" : "direct",
    );
  }, [config.webhook_url_ref]);

  const switchMode = (mode: string) => {
    setWebhookMode(mode);
    if (mode === "direct") onChange("webhook_url_ref", undefined);
    else onChange("webhook_url", undefined);
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
        showTemplateHint={false}
      />
    </div>
  );
}

export function ActionConfigFields({
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

  const secretOptions = (secrets ?? []).map((s) => ({
    label: s.name,
    value: s.name,
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
                No connections yet —{" "}
                <a
                  href="/dashboard/connections"
                  className="underline hover:text-amber-400"
                >
                  add one here
                </a>
                .
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

export function ActionCard({
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
  const setConfig = (key: string, value: unknown) => {
    const next = { ...action.config, [key]: value };
    if (value === undefined) delete next[key];
    onUpdate(index, { ...action, config: next });
  };

  return (
    <div className="rounded-xl border border-white/10 bg-[#141414] p-4">
      <div className="mb-3 flex items-center justify-between">
        <span className="text-xs font-medium text-zinc-400">
          Action {index + 1}
        </span>
        {canRemove && (
          <button
            type="button"
            onClick={() => onRemove(index)}
            className="text-xs text-red-400 transition-colors hover:text-red-300"
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
