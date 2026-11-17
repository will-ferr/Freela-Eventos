const dateTimeFmt = new Intl.DateTimeFormat('pt-BR', {
  dateStyle: 'short',
  timeStyle: 'medium',
})

export function formatDateTime(iso: string | Date): string {
  const d = typeof iso === 'string' ? new Date(iso) : iso
  return Number.isNaN(d.getTime()) ? '—' : dateTimeFmt.format(d)
}

export function formatLatency(ms: number): string {
  if (ms < 1) return '< 1 ms'
  if (ms < 1000) return `${Math.round(ms)} ms`
  return `${(ms / 1000).toFixed(1)} s`
}

/** Converte "1h2m3s" (formato Go) em "1h 2min". */
export function formatGoDuration(d: string): string {
  const match = /^(?:(\d+)h)?(?:(\d+)m)?(?:(\d+)s)?$/.exec(d)
  if (!match || d === '') return d
  const [, h = '0', m = '0', s = '0'] = match
  const hours = Number(h)
  const minutes = Number(m)
  if (hours >= 24) {
    const days = Math.floor(hours / 24)
    return `${days}d ${hours % 24}h`
  }
  if (hours > 0) return `${hours}h ${minutes}min`
  if (minutes > 0) return `${minutes}min`
  return `${Number(s)}s`
}

export function shortCommit(commit: string): string {
  return /^[0-9a-f]{7,40}$/i.test(commit) ? commit.slice(0, 7) : commit
}
