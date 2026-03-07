export interface User {
  id: string
  username: string
  email: string
  created_at: string
  updated_at: string
}

export interface AuthResponse {
  token: string
  user: User
}

export interface LoginRequest {
  email: string
  password: string
}

export interface RegisterRequest {
  username: string
  email: string
  password: string
}

export interface APIResponse<T = unknown> {
  success: boolean
  message?: string
  data?: T
}

export interface APIError {
  success: false
  error: string
  code: string
}
