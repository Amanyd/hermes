'use client'

import { useDeleteRelay, useRelays, useUpdateRelay } from '@/lib/queries'
import type { Relay } from '@/types/relay'
import Link from 'next/link'
import { toast } from 'sonner'

function StatusBadge({ active }: { active: boolean }) {
  return (
    <span className={`inline-flex items-center gap-1.5 rounded-full px-2.5 py-0.5 text-xs font-medium ${
      active ? 'bg-green-500/10 text-green-400' : 'bg-zinc-500/10 text-zinc-400'
    }`}>
      <span className={`h-1.5 w-1.5 rounded-full ${active ? 'bg-green-400' : 'bg-zinc-400'}`} />
      {active ? 'Active' : 'Inactive'}
    </span>
  )
}

function RelayRow({ relay }: { relay: Relay }) {
  const deleteMutation = useDeleteRelay()
  const updateMutation = useUpdateRelay(relay.id)

  const handleDelete = async () => {
    if (!confirm(`Delete relay "${relay.name}"?`)) return
    try {
      await deleteMutation.mutateAsync(relay.id)
      toast.success('Relay deleted')
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Failed to delete')
    }
  }

  const handleToggle = async () => {
    try {
      await updateMutation.mutateAsync({ is_active: !relay.is_active })
      toast.success(relay.is_active ? 'Relay deactivated' : 'Relay activated')
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Failed to update')
    }
  }

  return (
    <div className="flex items-center justify-between rounded-xl border border-white/10 bg-[#1a1a1a] px-5 py-4 transition-colors hover:border-white/20">
      <div className="flex items-center gap-4 min-w-0">
        <div className="min-w-0">
          <div className="flex items-center gap-3">
            <Link
              href={`/dashboard/relays/${relay.id}`}
              className="text-sm font-medium text-white hover:text-orange-400 transition-colors truncate"
            >
              {relay.name}
            </Link>
            <StatusBadge active={relay.is_active} />
          </div>
          {relay.description && (
            <p className="mt-0.5 text-xs text-zinc-500 truncate">{relay.description}</p>
          )}
          <div className="mt-1.5 flex items-center gap-1.5">
            <code className="rounded bg-white/5 px-2 py-0.5 text-xs text-zinc-400 font-mono">
              {relay.webhook_url}
            </code>
            <button
              type="button"
              onClick={() => {
                navigator.clipboard.writeText(relay.webhook_url)
                toast.success('Copied!')
              }}
              className="text-zinc-600 hover:text-zinc-300 transition-colors"
              title="Copy webhook URL"
            >
              <svg aria-hidden="true" className="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
                <path strokeLinecap="round" strokeLinejoin="round" d="M15.666 3.888A2.25 2.25 0 0 0 13.5 2.25h-3c-1.03 0-1.9.693-2.166 1.638m7.332 0c.055.194.084.4.084.612v0a.75.75 0 0 1-.75.75H9a.75.75 0 0 1-.75-.75v0c0-.212.03-.418.084-.612m7.332 0c.646.049 1.288.11 1.927.184 1.1.128 1.907 1.077 1.907 2.185V19.5a2.25 2.25 0 0 1-2.25 2.25H6.75A2.25 2.25 0 0 1 4.5 19.5V6.257c0-1.108.806-2.057 1.907-2.185a48.208 48.208 0 0 1 1.927-.184" />
              </svg>
            </button>
          </div>
        </div>
      </div>

      <div className="flex items-center gap-2 ml-4 shrink-0">
        <button
          type="button"
          onClick={handleToggle}
          disabled={updateMutation.isPending}
          className="rounded-lg border border-white/10 px-3 py-1.5 text-xs text-zinc-400 transition-colors hover:border-white/20 hover:text-white disabled:opacity-50"
        >
          {relay.is_active ? 'Deactivate' : 'Activate'}
        </button>
        <Link
          href={`/dashboard/relays/${relay.id}`}
          className="rounded-lg border border-white/10 px-3 py-1.5 text-xs text-zinc-400 transition-colors hover:border-white/20 hover:text-white"
        >
          View
        </Link>
        <button
          type="button"
          onClick={handleDelete}
          disabled={deleteMutation.isPending}
          className="rounded-lg border border-red-500/20 px-3 py-1.5 text-xs text-red-400 transition-colors hover:border-red-500/40 hover:text-red-300 disabled:opacity-50"
        >
          Delete
        </button>
      </div>
    </div>
  )
}

export default function RelaysPage() {
  const { data: relays, isLoading, isError } = useRelays()

  return (
    <div className="p-8">
      {/* Header */}
      <div className="flex items-center justify-between mb-8">
        <div>
          <h1 className="text-xl font-bold text-white">Relays</h1>
          <p className="mt-0.5 text-sm text-zinc-500">Manage your webhook automations</p>
        </div>
        <Link
          href="/dashboard/relays/new"
          className="flex items-center gap-2 rounded-lg bg-orange-500 px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-orange-600"
        >
          <svg aria-hidden="true" className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
            <path strokeLinecap="round" strokeLinejoin="round" d="M12 4.5v15m7.5-7.5h-15" />
          </svg>
          New Relay
        </Link>
      </div>

      {/* Content */}
      {isLoading && (
        <div className="space-y-3">
          {[...Array(3)].map((_, i) => (
            // biome-ignore lint/suspicious/noArrayIndexKey: skeleton loader
            <div key={i} className="h-20 animate-pulse rounded-xl bg-white/5" />
          ))}
        </div>
      )}

      {isError && (
        <div className="rounded-xl border border-red-500/20 bg-red-500/5 p-4 text-sm text-red-400">
          Failed to load relays. Make sure the backend is running.
        </div>
      )}

      {!isLoading && !isError && relays?.length === 0 && (
        <div className="flex flex-col items-center justify-center rounded-xl border border-dashed border-white/10 py-20 text-center">
          <svg aria-hidden="true" className="h-10 w-10 text-zinc-600 mb-3" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1}>
            <path strokeLinecap="round" strokeLinejoin="round" d="M7.5 21 3 16.5m0 0L7.5 12M3 16.5h13.5m0-13.5L21 7.5m0 0L16.5 12M21 7.5H7.5" />
          </svg>
          <p className="text-sm font-medium text-zinc-400">No relays yet</p>
          <p className="mt-1 text-xs text-zinc-600">Create your first relay to start routing webhooks</p>
          <Link
            href="/dashboard/relays/new"
            className="mt-4 rounded-lg bg-orange-500 px-4 py-2 text-sm font-medium text-white hover:bg-orange-600 transition-colors"
          >
            Create relay
          </Link>
        </div>
      )}

      {!isLoading && !isError && relays && relays.length > 0 && (
        <div className="space-y-3">
          {relays.map((relay) => (
            <RelayRow key={relay.id} relay={relay} />
          ))}
        </div>
      )}
    </div>
  )
}
