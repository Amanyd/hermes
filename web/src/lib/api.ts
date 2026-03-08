import type {
  APIError,
  APIResponse,
  AuthResponse,
  LoginRequest,
  RegisterRequest,
} from "@/types/auth";
import type {
  Connection,
  CreateRelayActionInput,
  CreateRelayRequest,
  CreateSecretRequest,
  Execution,
  ExecutionStep,
  Relay,
  RelayWithActions,
  Secret,
  UpdateRelayRequest,
} from "@/types/relay";

const API_BASE =
  process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:3000/api/v1";

function getToken(): string | null {
  if (typeof window === "undefined") return null;
  return localStorage.getItem("token");
}

async function apiFetch<T>(path: string, init?: RequestInit): Promise<T> {
  const token = getToken();
  const res = await fetch(`${API_BASE}${path}`, {
    ...init,
    headers: {
      "Content-Type": "application/json",
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
      ...init?.headers,
    },
  });

  const json = await res.json();

  if (!res.ok) {
    const err = json as APIError;
    throw new Error(err.error ?? "Something went wrong");
  }

  return json as T;
}

export async function login(data: LoginRequest): Promise<AuthResponse> {
  const res = await apiFetch<APIResponse<AuthResponse>>("/auth/login", {
    method: "POST",
    body: JSON.stringify(data),
  });
  return res.data!;
}

export async function register(data: RegisterRequest): Promise<AuthResponse> {
  const res = await apiFetch<APIResponse<AuthResponse>>("/auth/register", {
    method: "POST",
    body: JSON.stringify(data),
  });
  return res.data!;
}

export async function getRelays(): Promise<Relay[]> {
  const res = await apiFetch<APIResponse<Relay[]>>("/relays");
  return res.data ?? [];
}

export async function getRelay(id: string): Promise<RelayWithActions> {
  const res = await apiFetch<APIResponse<RelayWithActions>>(`/relays/${id}`);
  return res.data!;
}

export async function createRelay(
  data: CreateRelayRequest,
): Promise<RelayWithActions> {
  const res = await apiFetch<APIResponse<RelayWithActions>>("/relays", {
    method: "POST",
    body: JSON.stringify(data),
  });
  return res.data!;
}

export async function updateRelay(
  id: string,
  data: UpdateRelayRequest,
): Promise<Relay> {
  const res = await apiFetch<APIResponse<Relay>>(`/relays/${id}`, {
    method: "PUT",
    body: JSON.stringify(data),
  });
  return res.data!;
}

export async function deleteRelay(id: string): Promise<void> {
  await apiFetch(`/relays/${id}`, { method: "DELETE" });
}

export async function triggerRelay(
  id: string,
  payload: Record<string, unknown> = {},
): Promise<{ relay_id: string }> {
  const res = await apiFetch<APIResponse<{ relay_id: string }>>(
    `/relays/${id}/trigger`,
    {
      method: "POST",
      body: JSON.stringify(payload),
    },
  );
  return res.data!;
}

export async function getExecutions(
  id: string,
  limit = 50,
): Promise<Execution[]> {
  const res = await apiFetch<APIResponse<Execution[]>>(
    `/relays/${id}/executions?limit=${limit}`,
  );
  return res.data ?? [];
}

export async function getExecutionSteps(
  executionId: string,
): Promise<ExecutionStep[]> {
  const res = await apiFetch<APIResponse<ExecutionStep[]>>(
    `/executions/${executionId}/steps`,
  );
  return res.data ?? [];
}

export async function getSecrets(): Promise<Secret[]> {
  const res = await apiFetch<APIResponse<Secret[]>>("/secrets");
  return res.data ?? [];
}

export async function createSecret(data: CreateSecretRequest): Promise<Secret> {
  const res = await apiFetch<APIResponse<Secret>>("/secrets", {
    method: "POST",
    body: JSON.stringify(data),
  });
  return res.data!;
}

export async function deleteSecret(id: string): Promise<void> {
  await apiFetch(`/secrets/${id}`, { method: "DELETE" });
}

export async function getAvailableProviders(): Promise<{
  providers: string[];
}> {
  const res = await apiFetch<APIResponse<{ providers: string[] }>>(
    "/connections/providers",
  );
  return res.data ?? { providers: [] };
}

export async function getConnections(): Promise<Connection[]> {
  const res = await apiFetch<APIResponse<Connection[]>>("/connections");
  return res.data ?? [];
}

export async function deleteConnection(id: string): Promise<void> {
  await apiFetch(`/connections/${id}`, { method: "DELETE" });
}

export function getConnectURL(provider: string): string {
  return `${API_BASE}/connections/${provider}/connect`;
}

export async function updateRelayActions(
  relayId: string,
  actions: CreateRelayActionInput[],
): Promise<RelayWithActions> {
  const res = await apiFetch<APIResponse<RelayWithActions>>(
    `/relays/${relayId}/actions`,
    {
      method: "PUT",
      body: JSON.stringify({ actions }),
    },
  );
  return res.data!;
}
