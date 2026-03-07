'use client'

import { useSearchParams } from 'next/navigation'
import { Suspense, useEffect, useRef, useState } from 'react'
import { toast } from 'sonner'
import { getConnectURL } from '@/lib/api'
import {
  useAvailableProviders,
  useConnections,
  useDeleteConnection,
} from '@/lib/queries'
import type { Connection } from '@/types/relay'

const PROVIDER_META: Record<string, { label: string; color: string }> = {
  google: {
    label: 'Google',
    color: 'bg-red-500/10 text-red-400 border-red-500/20',
  },
  microsoft: {
    label: 'Microsoft',
    color: 'bg-blue-500/10 text-blue-400 border-blue-500/20',
  },
}

function GoogleIcon() {
  return (
    <svg
      aria-hidden="true"
      className="h-4 w-4"
      viewBox="0 0 24 24"
      fill="currentColor"
    >
      <path
        d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z"
        fill="#4285F4"
      />
      <path
        d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"
        fill="#34A853"
      />
      <path
        d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z"
        fill="#FBBC05"
      />
      <path
        d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z"
        fill="#EA4335"
      />
    </svg>
  )
}

function MicrosoftIcon() {
  return (
    <svg
      aria-hidden="true"
      className="h-4 w-4"
      viewBox="0 0 24 24"
      fill="currentColor"
    >
      <path d="M11.4 2H2v9.4h9.4V2z" fill="#F25022" />
      <path d="M22 2h-9.4v9.4H22V2z" fill="#7FBA00" />
      <path d="M11.4 12.6H2V22h9.4v-9.4z" fill="#00A4EF" />
      <path d="M22 12.6h-9.4V22H22v-9.4z" fill="#FFB900" />
    </svg>
  )
}

function ProviderIcon({ provider }: { provider: string }) {
  if (provider === 'google') return <GoogleIcon />
  if (provider === 'microsoft') return <MicrosoftIcon />
  return null
}

function DeleteModal({
  connection,
  onConfirm,
  onCancel,
  isPending,
}: {
  connection: Connection
  onConfirm: () => void
  onCancel: () => void
  isPending: boolean
}) {
  const meta = PROVIDER_META[connection.provider] ?? {
    label: connection.provider,
    color: '',
  }
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4">
      <div className="w-full max-w-sm rounded-xl border border-white/10 bg-[#1a1a1a] p-6 shadow-2xl">
        <h2 className="text-base font-semibold text-white">
          Remove connection?
        </h2>
        <p className="mt-2 text-sm text-zinc-400">
          This will remove your{' '}
          <span className="font-medium text-white">{meta.label}</span>{' '}
          connection for{' '}
          <span className="font-medium text-white">
            {connection.account_email}
          </span>
          . Actions using this connection will fail until reconnected.
        </p>
        <div className="mt-5 flex gap-3">
          <button
            type="button"
            onClick={onConfirm}
            disabled={isPending}
            className="flex-1 rounded-lg bg-red-500 px-4 py-2 text-sm font-semibold text-white transition-colors hover:bg-red-600 disabled:opacity-50"
          >
            {isPending ? 'Removing…' : 'Remove'}
          </button>
          <button
            type="button"
            onClick={onCancel}
            className="flex-1 rounded-lg border border-white/10 px-4 py-2 text-sm text-zinc-400 transition-colors hover:border-white/20 hover:text-white"
          >
            Cancel
          </button>
        </div>
      </div>
    </div>
  )
}

function ConnectionCard({ connection }: { connection: Connection }) {
  const [confirmDelete, setConfirmDelete] = useState(false)
  const deleteMutation = useDeleteConnection()
  const meta = PROVIDER_META[connection.provider] ?? {
    label: connection.provider,
    color: 'bg-zinc-500/10 text-zinc-400 border-zinc-500/20',
  }

  const handleDelete = () => {
    deleteMutation.mutate(connection.id, {
      onSuccess: () => {
        toast.success('Connection removed')
        setConfirmDelete(false)
      },
      onError: (err) => {
        toast.error(
          err instanceof Error ? err.message : 'Failed to remove connection',
        )
        setConfirmDelete(false)
      },
    })
  }

  return (
    <>
      <div className="flex items-center justify-between rounded-xl border border-white/10 bg-[#141414] px-4 py-3">
        <div className="flex items-center gap-3">
          <div className="flex h-8 w-8 items-center justify-center rounded-lg border border-white/10 bg-white/5">
            <ProviderIcon provider={connection.provider} />
          </div>
          <div>
            <div className="flex items-center gap-2">
              <span className="text-sm font-medium text-white">
                {connection.account_email}
              </span>
              <span
                className={`rounded-full border px-2 py-0.5 text-xs font-medium ${meta.color}`}
              >
                {meta.label}
              </span>
            </div>
            <p className="text-xs text-zinc-500">
              Connected{' '}
              {new Date(connection.created_at).toLocaleDateString('en-US', {
                month: 'short',
                day: 'numeric',
                year: 'numeric',
              })}
            </p>
          </div>
        </div>
        <button
          type="button"
          onClick={() => setConfirmDelete(true)}
          className="rounded-lg border border-white/10 px-3 py-1.5 text-xs text-zinc-400 transition-colors hover:border-red-500/30 hover:text-red-400"
        >
          Remove
        </button>
      </div>

      {confirmDelete && (
        <DeleteModal
          connection={connection}
          onConfirm={handleDelete}
          onCancel={() => setConfirmDelete(false)}
          isPending={deleteMutation.isPending}
        />
      )}
    </>
  )
}

function ConnectButton({ provider }: { provider: string }) {
  const meta = PROVIDER_META[provider] ?? { label: provider, color: '' }
  const token =
    typeof window !== 'undefined' ? localStorage.getItem('token') : null

  const handleConnect = () => {
    if (!token) {
      toast.error('Not authenticated')
      return
    }
    // Build URL — backend reads JWT from Authorization header via state codec
    // We pass token as query param since OAuth redirect can't carry headers
    const url = `${getConnectURL(provider)}?token=${encodeURIComponent(token)}`
    window.location.href = url
  }

  return (
    <button
      type="button"
      onClick={handleConnect}
      className="flex items-center gap-2.5 rounded-xl border border-white/10 bg-[#141414] px-4 py-3 text-sm text-zinc-300 transition-all hover:border-orange-500/30 hover:bg-orange-500/5 hover:text-white"
    >
      <div className="flex h-7 w-7 items-center justify-center rounded-lg border border-white/10 bg-white/5">
        <ProviderIcon provider={provider} />
      </div>
      Connect {meta.label}
      <svg
        aria-hidden="true"
        className="ml-auto h-3.5 w-3.5 text-zinc-600"
        fill="none"
        viewBox="0 0 24 24"
        stroke="currentColor"
        strokeWidth={2}
      >
        <path
          strokeLinecap="round"
          strokeLinejoin="round"
          d="M13.5 4.5 21 12m0 0-7.5 7.5M21 12H3"
        />
      </svg>
    </button>
  )
}

function ConnectionsContent() {
  const { data: providerData, isLoading: providersLoading } =
    useAvailableProviders()
  const { data: connections, isLoading: connectionsLoading } = useConnections()
  const searchParams = useSearchParams()
  const shownToast = useRef(false)

  useEffect(() => {
    if (shownToast.current) return
    if (searchParams.get('connected') === '1') {
      shownToast.current = true
      toast.success('Account connected successfully!')
    }
  }, [searchParams])

  const availableProviders = providerData?.providers ?? []

  return (
    <div className="p-8 max-w-2xl">
      <div className="mb-8">
        <h1 className="text-xl font-bold text-white">Connections</h1>
        <p className="mt-0.5 text-sm text-zinc-500">
          Connect your accounts to use email sending and other integrations in
          relay actions.
        </p>
      </div>

      {/* Add connection */}
      {availableProviders.length > 0 && (
        <section className="mb-8">
          <h2 className="mb-3 text-xs font-semibold uppercase tracking-wider text-zinc-500">
            Add connection
          </h2>
          {providersLoading ? (
            <div className="h-12 animate-pulse rounded-xl bg-white/5" />
          ) : (
            <div className="space-y-2">
              {availableProviders.map((provider) => (
                <ConnectButton key={provider} provider={provider} />
              ))}
            </div>
          )}
        </section>
      )}

      {/* Existing connections */}
      <section>
        <h2 className="mb-3 text-xs font-semibold uppercase tracking-wider text-zinc-500">
          Active connections{' '}
          {connections && connections.length > 0 && `(${connections.length})`}
        </h2>

        {connectionsLoading ? (
          <div className="space-y-2">
            {[0, 1].map((i) => (
              <div
                key={i}
                className="h-16 animate-pulse rounded-xl bg-white/5"
              />
            ))}
          </div>
        ) : !connections || connections.length === 0 ? (
          <div className="rounded-xl border border-dashed border-white/10 px-6 py-10 text-center">
            <svg
              aria-hidden="true"
              className="mx-auto mb-3 h-8 w-8 text-zinc-700"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
              strokeWidth={1.5}
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                d="M13.19 8.688a4.5 4.5 0 0 1 1.242 7.244l-4.5 4.5a4.5 4.5 0 0 1-6.364-6.364l1.757-1.757m13.35-.622 1.757-1.757a4.5 4.5 0 0 0-6.364-6.364l-4.5 4.5a4.5 4.5 0 0 0 1.242 7.244"
              />
            </svg>
            <p className="text-sm text-zinc-500">No connections yet</p>
            <p className="mt-0.5 text-xs text-zinc-600">
              Add a connection above to get started
            </p>
          </div>
        ) : (
          <div className="space-y-2">
            {connections.map((conn) => (
              <ConnectionCard key={conn.id} connection={conn} />
            ))}
          </div>
        )}
      </section>

      {/* Info banner — no providers configured */}
      {!providersLoading && availableProviders.length === 0 && (
        <div className="mt-6 rounded-xl border border-amber-500/20 bg-amber-500/5 px-4 py-3">
          <p className="text-sm font-medium text-amber-400">
            No OAuth providers configured
          </p>
          <p className="mt-0.5 text-xs text-amber-400/70">
            Set{' '}
            <code className="rounded bg-amber-500/10 px-1 font-mono">
              GOOGLE_CLIENT_ID
            </code>{' '}
            or{' '}
            <code className="rounded bg-amber-500/10 px-1 font-mono">
              MICROSOFT_CLIENT_ID
            </code>{' '}
            in the backend to enable connections.
          </p>
        </div>
      )}
    </div>
  )
}

export default function ConnectionsPage() {
  return (
    <Suspense>
      <ConnectionsContent />
    </Suspense>
  )
}
