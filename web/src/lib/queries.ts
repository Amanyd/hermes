import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  createRelay,
  createSecret,
  deleteConnection,
  deleteRelay,
  deleteSecret,
  getAvailableProviders,
  getConnections,
  getExecutionSteps,
  getExecutions,
  getRelay,
  getRelays,
  getSecrets,
  updateRelay,
  updateRelayActions,
} from "@/lib/api";
import type {
  CreateRelayActionInput,
  CreateRelayRequest,
  CreateSecretRequest,
  UpdateRelayRequest,
} from "@/types/relay";

export const queryKeys = {
  relays: ["relays"] as const,
  relay: (id: string) => ["relays", id] as const,
  relayExecutions: (id: string) => ["relays", id, "executions"] as const,
  executionSteps: (executionId: string) =>
    ["executions", executionId, "steps"] as const,
  secrets: ["secrets"] as const,
  connections: ["connections"] as const,
  connectionProviders: ["connections", "providers"] as const,
};

export function useRelays() {
  return useQuery({
    queryKey: queryKeys.relays,
    queryFn: getRelays,
  });
}

export function useRelay(id: string) {
  return useQuery({
    queryKey: queryKeys.relay(id),
    queryFn: () => getRelay(id),
    enabled: !!id,
  });
}

export function useExecutions(id: string) {
  return useQuery({
    queryKey: queryKeys.relayExecutions(id),
    queryFn: () => getExecutions(id),
    enabled: !!id,
    refetchInterval: 10_000,
  });
}

export function useExecutionSteps(executionId: string, enabled = true) {
  return useQuery({
    queryKey: queryKeys.executionSteps(executionId),
    queryFn: () => getExecutionSteps(executionId),
    enabled: !!executionId && enabled,
  });
}

export function useCreateRelay() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: CreateRelayRequest) => createRelay(data),
    onSuccess: () => qc.invalidateQueries({ queryKey: queryKeys.relays }),
  });
}

export function useUpdateRelay(id: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: UpdateRelayRequest) => updateRelay(id, data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: queryKeys.relays });
      qc.invalidateQueries({ queryKey: queryKeys.relay(id) });
    },
  });
}

export function useDeleteRelay() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => deleteRelay(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: queryKeys.relays }),
  });
}

export function useSecrets() {
  return useQuery({
    queryKey: queryKeys.secrets,
    queryFn: getSecrets,
  });
}

export function useCreateSecret() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: CreateSecretRequest) => createSecret(data),
    onSuccess: () => qc.invalidateQueries({ queryKey: queryKeys.secrets }),
  });
}

export function useDeleteSecret() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => deleteSecret(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: queryKeys.secrets }),
  });
}

export function useAvailableProviders() {
  return useQuery({
    queryKey: queryKeys.connectionProviders,
    queryFn: getAvailableProviders,
    staleTime: 60_000,
  });
}

export function useConnections() {
  return useQuery({
    queryKey: queryKeys.connections,
    queryFn: getConnections,
    staleTime: 30_000,
  });
}

export function useDeleteConnection() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => deleteConnection(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: queryKeys.connections }),
  });
}

export function useUpdateRelayActions(id: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (actions: CreateRelayActionInput[]) =>
      updateRelayActions(id, actions),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: queryKeys.relay(id) });
      qc.invalidateQueries({ queryKey: queryKeys.relayExecutions(id) });
    },
  });
}
