'use client'

import { useDeleteRelay, useRelay, useRelayLogs, useUpdateRelay } from '@/lib/queries'
import { ACTION_LABELS } from '@/types/relay'
import { useParams, useRouter } from 'next/navigation'
import { useState } from 'react'
import { toast } from 'sonner'

function StatusBadge({ active }: { active: boolean }) {
  return (
    <span
      className={`inline-flex items-center gap-1.5 rounded-full px-2.5 py-0.5 text-xs font-medium ${
        active ? 'bg-green-500/10 text-green-400' : 'bg-zinc-800 text-zinc-500'
      }`}
    >
      <span className={`h-1.5 w-1.5 rounded-full ${active ? 'bg-green-400' : 'bg-zinc-500'}`} />
      {active ? 'Active' : 'Inactive'}
    </span>
  )
}

function LogStatusBadge({ status }: { status: string }) {
  const map: Record<string, string> = {
    success: 'bg-green-500/10 text-green-400',
    failure: 'bg-red-500/10 text-red-400',
    failed: 'bg-red-500/10 text-red-400',
    pending: 'bg-yellow-500/10 text-yellow-400',
  }
  return (
    <span className={`inline-flex rounded-full px-2.5 py-0.5 text-xs font-medium ${map[status] ?? 'bg-zinc-800 text-zinc-400'}`}>
      {status}
    </span>
  )
}

export default function RelayDetailPage() {
  const { id } = useParams<{ id: string }>()
  const router = useRouter()
  const { data: relay, isLoading, isError } = useRelay(id)
  const { data: logs } = useRelayLogs(id)
  const updateMutation = useUpdateRelay(id)
  const deleteMutation = useDeleteRelay()
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false)
  const [copied, setCopied] = useState(false)

  const webhookUrl = relay?.webhook_url ?? ''

  const copyWebhook = async () => {
    await navigator.clipboard.writeText(webhookUrl)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  const toggleActive = async () => {
    if (!relay) return
    try {
      await updateMutation.mutateAsync({ is_active: !relay.is_active })
      toast.success(relay.is_active ? 'Relay deactivated' : 'Relay activated')
    } catch {
      toast.error('Failed to update relay')
    }
  }

  const handleDelete = async () => {
    try {
      await deleteMutation.mutateAsync(id)
      toast.success('Relay deleted')
      router.push('/dashboard/relays')
    } catch {
      toast.error('Failed to delete relay')
    }
  }

  if (isLoading) {
    return (
      <div className="p-8 space-y-4 max-w-3xl">
        {[...Array(4)].map((_, i) => (
          // biome-ignore lint/suspicious/noArrayIndexKey: static skeleton
          <div key={i} className="h-12 rounded-lg bg-white/5 animate-pulse" />
        ))}
      </div>
    )
  }

  if (isError || !relay) {
    return (
      <div className="p-8">
        <p className="text-red-400 text-sm">Failed to load relay.</p>
      </div>
    )
  }

  return (
    <div className="p-8 max-w-3xl space-y-8">
      {/* Header */}
      <div className="flex items-start justify-between gap-4">
        <div>
          <div className="flex items-center gap-3 mb-1">
            <h1 className="text-xl font-bold text-white">{relay.name}</h1>
            <StatusBadge active={relay.is_active} />
          </div>
          {relay.description && <p className="text-sm text-zinc-500">{relay.description}</p>}
        </div>
        <div className="flex items-center gap-2 shrink-0">
          <button
            type="button"
            onClick={toggleActive}
            disabled={updateMutation.isPending}
            className="rounded-lg border border-white/10 px-3.5 py-1.5 text-xs font-medium text-zinc-300 transition-colors hover:border-white/20 hover:text-white disabled:opacity-50"
          >
            {relay.is_active ? 'Deactivate' : 'Activate'}
          </button>
          <button
            type="button"
            onClick={() => setShowDeleteConfirm(true)}
            className="rounded-lg border border-red-500/20 px-3.5 py-1.5 text-xs font-medium text-red-400 transition-colors hover:border-red-500/40 hover:text-red-300"
          >
            Delete
          </button>
        </div>
      </div>

      {/* Webhook URL */}
      <div className="rounded-xl border border-white/10 bg-[#1a1a1a] p-4">
        <p className="mb-2 text-xs font-medium text-zinc-400">Webhook URL</p>
        <div className="flex items-center gap-2">
          <code className="flex-1 truncate rounded-lg bg-black/30 px-3 py-2 text-xs text-zinc-300 font-mono">
            {webhookUrl}
          </code>
          <button
            type="button"
            onClick={copyWebhook}
            className="shrink-0 rounded-lg bg-white/5 px-3 py-2 text-xs text-zinc-400 transition-colors hover:bg-white/10 hover:text-white"
          >
            {copied ? 'Copied!' : 'Copy'}
          </button>
        </div>
      </div>

      {/* Actions */}
      {relay.actions && relay.actions.length > 0 && (
        <div>
          <h2 className="mb-3 text-sm font-medium text-zinc-300">Actions ({relay.actions.length})</h2>
          <div className="space-y-2">
            {relay.actions.map((action, i) => (
              <div
                key={action.id}
                className="flex items-center gap-3 rounded-xl border border-white/10 bg-[#1a1a1a] px-4 py-3"
              >
                <span className="flex h-6 w-6 items-center justify-center rounded-full bg-orange-500/10 text-xs font-bold text-orange-400">
                  {i + 1}
                </span>
                <span className="text-sm text-white">
                  {ACTION_LABELS[action.action_type as keyof typeof ACTION_LABELS] ?? action.action_type}
                </span>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Execution Logs */}
      <div>
        <h2 className="mb-3 text-sm font-medium text-zinc-300">Execution Logs</h2>
        {!logs || logs.length === 0 ? (
          <div className="rounded-xl border border-white/10 bg-[#1a1a1a] px-4 py-8 text-center">
            <p className="text-sm text-zinc-500">No executions yet — send a request to your webhook URL to see logs here.</p>
          </div>
        ) : (
          <div className="rounded-xl border border-white/10 bg-[#1a1a1a] divide-y divide-white/5 overflow-hidden">
            <div className="grid grid-cols-[1fr_120px_160px] gap-4 px-4 py-2.5 text-xs font-medium text-zinc-500">
              <span>ID</span>
              <span>Status</span>
              <span>Created</span>
            </div>
            {logs.map((log) => (
              <div key={log.id} className="grid grid-cols-[1fr_120px_160px] gap-4 px-4 py-3 text-sm hover:bg-white/2 transition-colors">
                <span className="font-mono text-xs text-zinc-400 truncate">{log.id}</span>
                <LogStatusBadge status={log.status} />
                <span className="text-xs text-zinc-500">
                  {new Date(log.executed_at).toLocaleString()}
                </span>
              </div>
            ))}
          </div>
        )}
      </div>

      {/* Delete confirm modal */}
      {showDeleteConfirm && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/70">
          <div className="w-full max-w-sm rounded-2xl border border-white/10 bg-[#1a1a1a] p-6">
            <h3 className="mb-1 text-base font-semibold text-white">Delete relay?</h3>
            <p className="mb-5 text-sm text-zinc-500">
              This will permanently delete <span className="text-white font-medium">{relay.name}</span> and all its logs.
            </p>
            <div className="flex gap-3">
              <button
                type="button"
                onClick={handleDelete}
                disabled={deleteMutation.isPending}
                className="flex-1 rounded-lg bg-red-500 py-2 text-sm font-medium text-white transition-colors hover:bg-red-600 disabled:opacity-50"
              >
                {deleteMutation.isPending ? 'Deleting…' : 'Delete'}
              </button>
              <button
                type="button"
                onClick={() => setShowDeleteConfirm(false)}
                className="flex-1 rounded-lg border border-white/10 py-2 text-sm text-zinc-300 transition-colors hover:border-white/20 hover:text-white"
              >
                Cancel
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
