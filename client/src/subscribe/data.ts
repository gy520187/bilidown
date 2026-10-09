import { ResJSON } from '../mixin'

export type SubscriptionSource = {
    id: number
    type: 'season' | 'series' | 'fav' | 'space' | 'bangumi'
    mid: string
    resourceId: string
    title: string
    cover: string
    owner: string
    enabled: boolean
    downloadType: 'merge' | 'audio' | 'video'
    format: number
    cron: string
    sectionIds: string[]
    lastCheckAt: string
    lastCheckStatus: string
    lastError: string
    createdAt: string
    itemCount: number
}

export type PreviewSection = {
    id: string
    title: string
    count: number
}

export type SubscriptionPreview = {
    source: SubscriptionSource
    sections: PreviewSection[]
    total: number
}

export const typeLabel: Record<SubscriptionSource['type'], string> = {
    season: '合集',
    series: '系列',
    fav: '收藏夹',
    space: 'UP 空间',
    bangumi: '番剧',
}

export const statusLabel: Record<string, string> = {
    success: '成功',
    error: '失败',
    need_login: '需要重新登录',
}

export const getSubscriptions = async (): Promise<SubscriptionSource[]> => {
    const res = await fetch('/api/getSubscriptions').then(r => r.json()) as ResJSON<SubscriptionSource[]>
    if (!res.success) throw new Error(res.message)
    return res.data || []
}

export const previewSubscription = async (url: string): Promise<SubscriptionPreview> => {
    const res = await fetch('/api/previewSubscription', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ url }),
    }).then(r => r.json()) as ResJSON<SubscriptionPreview>
    if (!res.success) throw new Error(res.message)
    return res.data
}

export const createSubscription = async (payload: {
    url: string
    downloadType: SubscriptionSource['downloadType']
    format: number
    cron: string
    sectionIds: string[]
}): Promise<SubscriptionSource> => {
    const res = await fetch('/api/createSubscription', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload),
    }).then(r => r.json()) as ResJSON<SubscriptionSource>
    if (!res.success) throw new Error(res.message)
    return res.data
}

export const updateSubscription = async (payload: {
    id: number
    enabled?: boolean
    downloadType?: SubscriptionSource['downloadType']
    format?: number
    cron?: string
}) => {
    const res = await fetch('/api/updateSubscription', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload),
    }).then(r => r.json()) as ResJSON
    if (!res.success) throw new Error(res.message)
}

export const deleteSubscription = async (id: number) => {
    const res = await fetch(`/api/deleteSubscription?id=${id}`, { method: 'POST' }).then(r => r.json()) as ResJSON
    if (!res.success) throw new Error(res.message)
}

export const runSubscriptionCheck = async (id?: number) => {
    const query = id ? `?id=${id}` : ''
    const res = await fetch(`/api/runSubscriptionCheck${query}`, { method: 'POST' }).then(r => r.json()) as ResJSON
    if (!res.success) throw new Error(res.message)
}
