import van, { State } from 'vanjs-core'
import { Route, goto } from 'vanjs-router'
import { checkLogin, GLOBAL_HAS_LOGIN, VanComponent } from '../mixin'
import { LoadingBox } from '../view'
import {
    createSubscription,
    deleteSubscription,
    getSubscriptions,
    previewSubscription,
    runSubscriptionCheck,
    statusLabel,
    typeLabel,
    updateSubscription,
    type PreviewSection,
    type SubscriptionPreview,
    type SubscriptionSource
} from './data'

const { button, div, input, option, select, span, img, label } = van.tags

const formatOptions: { value: number, label: string }[] = [
    { value: 127, label: '8K' },
    { value: 126, label: '杜比视界' },
    { value: 125, label: 'HDR' },
    { value: 120, label: '4K' },
    { value: 116, label: '1080P60' },
    { value: 112, label: '1080P+' },
    { value: 80, label: '1080P' },
    { value: 64, label: '720P' },
    { value: 32, label: '480P' },
    { value: 16, label: '360P' },
]

const DEFAULT_CRON = '0 8 * * *'

const sectionLabel = (item: SubscriptionSource) => {
    const ids = item.sectionIds || []
    if (ids.length == 0) {
        return item.type == 'bangumi' ? '正片' : '全部稿件'
    }
    return ids.map(id => id == 'main' ? '正片' : id == 'all' ? '全部稿件' : id).join(', ')
}

export class SubscribeRoute implements VanComponent {
    element: HTMLElement
    loading = van.state(true)
    saving = van.state(false)
    checking = van.state(false)
    list: State<SubscriptionSource[]> = van.state([])
    urlValue = van.state('')
    downloadType = van.state<'merge' | 'audio' | 'video'>('merge')
    format = van.state(80)
    cronValue = van.state(DEFAULT_CRON)
    preview: State<SubscriptionPreview | null> = van.state(null)
    selectedSections: State<string[]> = van.state([])

    constructor() {
        this.element = this.Root()
    }

    async refresh() {
        this.list.val = await getSubscriptions()
    }

    Root() {
        const _that = this
        return Route({
            rule: 'subscribe',
            Loader() {
                return div(
                    () => _that.loading.val ? LoadingBox() : '',
                    () => _that.loading.val ? '' : div({ class: 'vstack gap-4' },
                        _that.AddBar(),
                        _that.List()
                    )
                )
            },
            async onFirst() {
                if (!await checkLogin()) return
                _that.loading.val = true
                try {
                    await _that.refresh()
                } catch (error) {
                    if (error instanceof Error) alert(error.message)
                } finally {
                    setTimeout(() => { _that.loading.val = false }, 200)
                }
            },
            onLoad() {
                if (!GLOBAL_HAS_LOGIN.val) return goto('login')
            }
        })
    }

    defaultSectionIds(sections: PreviewSection[]) {
        const main = sections.find(item => item.id == 'main')
        if (main) return [main.id]
        return sections.map(item => item.id)
    }

    toggleSection(id: string) {
        const current = this.selectedSections.val
        if (current.includes(id)) {
            this.selectedSections.val = current.filter(item => item != id)
            return
        }
        this.selectedSections.val = [...current, id]
    }

    AddBar() {
        const _that = this
        return div({ class: 'card card-body rounded-4 vstack gap-3' },
            div({ class: 'hstack gap-3 justify-content-between flex-wrap' },
                div({ class: 'fw-bold' }, '添加订阅'),
                button({
                    class: 'btn btn-outline-primary btn-sm',
                    disabled: _that.checking,
                    async onclick() {
                        _that.checking.val = true
                        try {
                            await runSubscriptionCheck()
                            alert('已开始检查全部订阅')
                            setTimeout(() => { _that.refresh().catch(() => { }) }, 1500)
                        } catch (error) {
                            if (error instanceof Error) alert(error.message)
                        } finally {
                            _that.checking.val = false
                        }
                    }
                }, () => _that.checking.val ? '检查中' : '立即检查全部')
            ),
            div({ class: 'text-secondary small' }, '支持合集、系列、收藏夹、空间、番剧，以及合集中的视频或 b23 短链。Cron 按北京时间，默认每天 08:00。'),
            div({ class: 'input-group' },
                input({
                    class: 'form-control',
                    placeholder: '粘贴订阅链接，支持 b23.tv',
                    value: _that.urlValue,
                    oninput: (e: Event) => _that.urlValue.val = (e.target as HTMLInputElement).value
                }),
                button({
                    class: 'btn btn-outline-secondary',
                    disabled: _that.saving,
                    async onclick() {
                        const url = _that.urlValue.val.trim()
                        if (!url) return alert('请输入订阅链接')
                        _that.saving.val = true
                        try {
                            const preview = await previewSubscription(url)
                            _that.preview.val = preview
                            _that.selectedSections.val = _that.defaultSectionIds(preview.sections || [])
                        } catch (error) {
                            _that.preview.val = null
                            if (error instanceof Error) alert(error.message)
                        } finally {
                            _that.saving.val = false
                        }
                    }
                }, '解析')
            ),
            () => {
                const preview = _that.preview.val
                if (!preview) return ''
                return div({ class: 'vstack gap-2' },
                    div({ class: 'fw-bold' }, preview.source.title || '未命名订阅'),
                    div({ class: 'small text-secondary' },
                        `${typeLabel[preview.source.type]} · 共 ${preview.total} 条`
                    ),
                    div({ class: 'vstack gap-1' },
                        (preview.sections || []).map(section => label({ class: 'form-check' },
                            input({
                                class: 'form-check-input',
                                type: 'checkbox',
                                checked: () => _that.selectedSections.val.includes(section.id),
                                oninput: () => _that.toggleSection(section.id)
                            }),
                            span({ class: 'form-check-label' }, `${section.title}（${section.count}）`)
                        ))
                    )
                )
            },
            div({ class: 'hstack gap-3 flex-wrap' },
                select({
                    class: 'form-select',
                    style: 'max-width: 160px',
                    value: _that.downloadType,
                    oninput: (e: Event) => _that.downloadType.val = (e.target as HTMLSelectElement).value as 'merge' | 'audio' | 'video'
                },
                    option({ value: 'merge' }, '音视频合并'),
                    option({ value: 'audio' }, '仅音频'),
                    option({ value: 'video' }, '仅视频'),
                ),
                select({
                    class: 'form-select',
                    style: 'max-width: 160px',
                    value: () => String(_that.format.val),
                    oninput: (e: Event) => _that.format.val = Number((e.target as HTMLSelectElement).value)
                },
                    formatOptions.map(item => option({ value: String(item.value) }, item.label))
                ),
                input({
                    class: 'form-control',
                    style: 'max-width: 180px',
                    placeholder: 'cron，默认 0 8 * * *',
                    value: _that.cronValue,
                    oninput: (e: Event) => _that.cronValue.val = (e.target as HTMLInputElement).value
                }),
                button({
                    class: 'btn btn-primary',
                    disabled: _that.saving,
                    async onclick() {
                        const url = _that.urlValue.val.trim()
                        if (!url) return alert('请输入订阅链接')
                        if (!_that.preview.val) return alert('请先解析订阅链接')
                        if (_that.selectedSections.val.length == 0) return alert('请至少选择一个分区')
                        _that.saving.val = true
                        try {
                            await createSubscription({
                                url,
                                downloadType: _that.downloadType.val,
                                format: _that.format.val,
                                cron: _that.cronValue.val.trim() || DEFAULT_CRON,
                                sectionIds: _that.selectedSections.val
                            })
                            _that.urlValue.val = ''
                            _that.cronValue.val = DEFAULT_CRON
                            _that.preview.val = null
                            _that.selectedSections.val = []
                            await _that.refresh()
                        } catch (error) {
                            if (error instanceof Error) alert(error.message)
                        } finally {
                            _that.saving.val = false
                        }
                    }
                }, '添加订阅')
            )
        )
    }

    List() {
        const _that = this
        return () => div({ class: 'vstack gap-3' },
            _that.list.val.length == 0
                ? div({ class: 'text-secondary' }, '还没有订阅源')
                : _that.list.val.map(item => _that.ItemCard(item))
        )
    }

    ItemCard(item: SubscriptionSource) {
        const _that = this
        const cronDraft = van.state(item.cron || DEFAULT_CRON)
        const statusText = statusLabel[item.lastCheckStatus] || (item.lastCheckStatus || '未检查')
        const sectionText = sectionLabel(item)
        return div({ class: 'card card-body rounded-4' },
            div({ class: 'hstack gap-3 align-items-start' },
                item.cover ? img({
                    src: item.cover,
                    class: 'rounded-3',
                    style: 'width: 96px; height: 64px; object-fit: cover;',
                    referrerpolicy: 'no-referrer'
                }) : div({ class: 'bg-light rounded-3', style: 'width: 96px; height: 64px;' }),
                div({ class: 'vstack gap-1 flex-fill' },
                    div({ class: 'fw-bold' }, item.title || `${typeLabel[item.type]} ${item.resourceId}`),
                    div({ class: 'small text-secondary' },
                        `${typeLabel[item.type]} · ${item.owner || '未知'} · 已记录 ${item.itemCount} 条 · 分区 ${sectionText}`
                    ),
                    div({ class: 'small' },
                        span({ class: item.lastCheckStatus == 'error' || item.lastCheckStatus == 'need_login' ? 'text-danger' : 'text-secondary' },
                            `最近检查：${item.lastCheckAt || '尚未检查'} · ${statusText}`
                        )
                    ),
                    item.lastError ? div({ class: 'small text-danger' }, item.lastError) : '',
                    div({ class: 'input-group input-group-sm mt-1', style: 'max-width: 420px' },
                        input({
                            class: 'form-control',
                            value: cronDraft,
                            oninput: (e: Event) => cronDraft.val = (e.target as HTMLInputElement).value
                        }),
                        button({
                            class: 'btn btn-outline-success',
                            async onclick() {
                                const expr = cronDraft.val.trim()
                                if (!expr) return alert('请输入 cron 表达式')
                                try {
                                    await updateSubscription({ id: item.id, cron: expr })
                                    await _that.refresh()
                                } catch (error) {
                                    if (error instanceof Error) alert(error.message)
                                }
                            }
                        }, '保存计划'),
                        button({
                            class: 'btn btn-outline-primary',
                            async onclick() {
                                try {
                                    await runSubscriptionCheck(item.id)
                                    alert('已开始检查该订阅')
                                    setTimeout(() => { _that.refresh().catch(() => { }) }, 1500)
                                } catch (error) {
                                    if (error instanceof Error) alert(error.message)
                                }
                            }
                        }, '立即检查')
                    )
                ),
                div({ class: 'vstack gap-2', style: 'min-width: 96px' },
                    button({
                        class: `btn btn-sm ${item.enabled ? 'btn-outline-secondary' : 'btn-outline-success'}`,
                        async onclick() {
                            try {
                                await updateSubscription({ id: item.id, enabled: !item.enabled })
                                await _that.refresh()
                            } catch (error) {
                                if (error instanceof Error) alert(error.message)
                            }
                        }
                    }, item.enabled ? '停用' : '启用'),
                    button({
                        class: 'btn btn-sm btn-outline-danger',
                        async onclick() {
                            if (!confirm('确定删除该订阅源？已下载文件会保留。')) return
                            try {
                                await deleteSubscription(item.id)
                                await _that.refresh()
                            } catch (error) {
                                if (error instanceof Error) alert(error.message)
                            }
                        }
                    }, '删除')
                )
            )
        )
    }
}

export default () => new SubscribeRoute().element
