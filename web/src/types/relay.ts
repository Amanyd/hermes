export interface Relay {
  id: string
  user_id: string
  name: string
  description: string
  webhook_path: string
  webhook_url: string
  is_active: boolean
  created_at: string
  updated_at: string
}

export interface RelayAction {
  id: string
  relay_id: string
  action_type: string
  config: Record<string, unknown>
  order_index: number
  created_at: string
  updated_at: string
}

export interface RelayWithActions extends Relay {
  actions: RelayAction[]
}

export interface ExecutionLog {
  id: string
  relay_id: string
  status: 'success' | 'failed'
  payload?: Record<string, unknown>
  error_message?: string
  executed_at: string
}

export interface Secret {
  id: string
  user_id: string
  name: string
  created_at: string
  updated_at: string
}

export interface CreateRelayActionInput {
  action_type: string
  config: Record<string, unknown>
  order_index: number
}

export interface CreateRelayRequest {
  name: string
  description?: string
  actions: CreateRelayActionInput[]
}

export interface UpdateRelayRequest {
  name?: string
  description?: string
  is_active?: boolean
}

export interface CreateSecretRequest {
  name: string
  value: string
}

export interface Connection {
  id: string
  user_id: string
  provider: 'google' | 'microsoft'
  account_email: string
  scopes: string
  created_at: string
  updated_at: string
}

export interface UpdateRelayActionsRequest {
  actions: CreateRelayActionInput[]
}

export const ACTION_TYPES = [
  'discord_send',
  'slack_send',
  'http_request',
  'email_send',
  'debug_log',
] as const

export type ActionType = (typeof ACTION_TYPES)[number]

export const ACTION_LABELS: Record<ActionType, string> = {
  discord_send: 'Discord',
  slack_send: 'Slack',
  http_request: 'HTTP Request',
  email_send: 'Email',
  debug_log: 'Debug Log',
}
