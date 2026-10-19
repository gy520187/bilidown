import van, { State } from 'vanjs-core'
import { Modal } from 'bootstrap'
import { VanComponent } from '../mixin'
import {
    buildSubscriptionURL,
    previewSubscription,
    typeLabel,
    updateSubscription,
    type PreviewSection,
    type SubscriptionPreview,
    type SubscriptionSource
} from './data'

const { button, div, input, label, option, select, span } = van.tags

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

export class ReconfigureModal implements VanComponent {
    element: HTMLElement
    modal: Modal
    preview: State<SubscriptionPreview | null> = van.state(null)
    selectedSections: State<string[]> = van.state([])
    downloadType = van.state<'merge' | 'audio' | 'video'>('merge')
    format = van.state(127)
    cronValue = van.state('')
    saving = van.state(false)
    loading = van.state(false)

    constructor(public item: SubscriptionSource, private onSaved: () => void) {
        this.element = this.Root()
        this.element.addEventListener('show.bs.modal', () => { this.loadPreview() })
        van.add(document.body, this.element)
        this.modal = new Modal(this.element)
    }

    show() {
        this.modal.show()
    }

    hide() {
        this.modal.hide()
    }

    async loadPreview() {
        this.loading.val = true
        this.preview.val = null
        this.downloadType.val = this.item.downloadType || 'merge'
        this.format.val = this.item.format || 127
        this.cronValue.val = this.item.cron || ''
        this.selectedSections.val = [...(this.item.sectionIds || [])]
        if (this.item.type != 'bangumi' && this.selectedSections.val.length == 0) {
            this.selectedSections.val = ['all']
        }
        try {
            this.preview.val = await previewSubscription(buildSubscriptionURL(this.item))
        } catch (error) {
            const fallbackSections: PreviewSection[] = (this.item.sectionIds || []).map(id => ({
                id,
                title: id == 'main' ? '正片' : id == 'all' ? '全部稿件' : id,
                count: 0,
            }))
            this.preview.val = {
                source: this.item,
                sections: fallbackSections,
                total: this.item.itemCount,
            }
            alert(error instanceof Error ? error.message : '重新解析失败，已回退到当前配置')
        } finally {
            this.loading.val = false
        }
    }

    async save() {
        if (this.item.type == 'bangumi' && this.selectedSections.val.length == 0) {
            alert('请至少选择一个分区')
            return
        }
        this.saving.val = true
        try {
            await updateSubscription({
                id: this.item.id,
                downloadType: this.downloadType.val,
                format: this.format.val,
                cron: this.cronValue.val.trim(),
                sectionIds: this.selectedSections.val,
            })
            this.hide()
            this.onSaved()
        } catch (error) {
            alert(error instanceof Error ? error.message : '保存失败')
        } finally {
            this.saving.val = false
        }
    }

    toggleSection(id: string) {
        const current = this.selectedSections.val
        if (current.includes(id)) {
            this.selectedSections.val = current.filter(item => item != id)
            return
        }
        this.selectedSections.val = [...current, id]
    }

    Root() {
        const _that = this
        return div({ class: 'modal fade', tabIndex: -1 },
            div({ class: 'modal-dialog modal-lg modal-dialog-centered modal-dialog-scrollable' },
                div({ class: 'modal-content' },
                    div({ class: 'modal-header' },
                        div({ class: 'h5 modal-title' }, '重新配置订阅'),
                        button({ class: 'btn-close', 'data-bs-dismiss': 'modal' })
                    ),
                    div({ class: 'modal-body vstack gap-3' },
                        () => {
                            const preview = _that.preview.val
                            return div({ class: 'vstack gap-1' },
                                div({ class: 'fw-bold' }, preview ? preview.source.title || '未命名订阅' : (_that.item.title || '未命名订阅')),
                                div({ class: 'small text-secondary' },
                                    preview
                                        ? `${typeLabel[preview.source.type]} · 共 ${preview.total} 条`
                                        : `${typeLabel[_that.item.type]} · 已记录 ${_that.item.itemCount} 条`
                                ),
                                div({ class: 'small text-secondary text-break' },
                                    '链接：', buildSubscriptionURL(_that.item)
                                )
                            )
                        },
                        () => _that.loading.val
                            ? div({ class: 'text-center text-secondary py-3' }, '正在加载最新分区...')
                            : div({ class: 'vstack gap-1' },
                                (_that.preview.val?.sections || []).map(section => label({ class: 'form-check' },
                                    input({
                                        class: 'form-check-input',
                                        type: 'checkbox',
                                        checked: () => _that.selectedSections.val.includes(section.id),
                                        oninput: () => _that.toggleSection(section.id)
                                    }),
                                    span({ class: 'form-check-label' }, `${section.title}（${section.count}）`)
                                ))
                            ),
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
                            })
                        )
                    ),
                    div({ class: 'modal-footer' },
                        button({ class: 'btn btn-secondary', 'data-bs-dismiss': 'modal' }, '取消'),
                        button({
                            class: 'btn btn-primary',
                            disabled: () => _that.saving.val,
                            onclick: () => { _that.save() }
                        }, () => _that.saving.val ? '保存中...' : '保存')
                    )
                )
            )
        )
    }
}