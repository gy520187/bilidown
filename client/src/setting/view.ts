import van from 'vanjs-core'
import { SettingRoute } from '.'
import { saveFields, pushTest } from './data'

const { a, button, div, input, label, option, select, span } = van.tags

export const SaveFolderSetting = (route: SettingRoute) => {
    const saveFolder = route.fields.download_folder
    const folderPickerDisabled = van.state(false)
    const buttonText = '保存'

    return div({ class: 'input-group' },
        div({ class: 'input-group-text' }, '下载目录'),
        input({
            class: 'form-control',
            value: saveFolder,
            oninput: event => saveFolder.val = event.target.value,
        }),
        button({
            class: 'btn btn-success', onclick() {
                folderPickerDisabled.val = true
                saveFields([
                    ['download_folder', saveFolder.val]
                ]).then(message => {
                    alert(message)
                }).catch(error => {
                    if (error instanceof Error) alert(error.message)
                }).finally(() => {
                    folderPickerDisabled.val = false
                })
            }, disabled: folderPickerDisabled
        }, buttonText)
    )
}
export const PushSetting = (route: SettingRoute) => {
    const enabled = route.fields.push_enabled
    const url = route.fields.push_url
    const token = route.fields.push_token
    const targetType = route.fields.push_target_type
    const targetID = route.fields.push_target_id

    const saving = van.state(false)
    const testing = van.state(false)

    const currentTargetType = () => targetType.val || 'private'

    const payload = () => ({
        url: url.val.trim(),
        token: token.val,
        target_type: currentTargetType(),
        target_id: targetID.val.trim()
    })

    const save = () => {
        if (enabled.val === '1') {
            if (!url.val.trim()) { alert('启用推送时，NapCat HTTP 地址不能为空'); return }
            if (!targetID.val.trim()) { alert('启用推送时，推送目标不能为空'); return }
        }
        saving.val = true
        saveFields([
            ['push_enabled', enabled.val === '1' ? '1' : '0'],
            ['push_channel', 'napcat'],
            ['push_url', url.val.trim()],
            ['push_token', token.val],
            ['push_target_type', currentTargetType()],
            ['push_target_id', targetID.val.trim()]
        ]).then(message => {
            alert(message)
        }).catch(error => {
            if (error instanceof Error) alert(error.message)
        }).finally(() => {
            saving.val = false
        })
    }

    const test = () => {
        if (!url.val.trim() || !targetID.val.trim()) {
            alert('请先填写 NapCat HTTP 地址和推送目标')
            return
        }
        testing.val = true
        pushTest(payload()).then(message => {
            alert(message)
        }).catch(error => {
            if (error instanceof Error) alert(error.message)
        }).finally(() => {
            testing.val = false
        })
    }

    return div({ class: 'card' },
        div({ class: 'card-header' }, '消息推送'),
        div({ class: 'card-body vstack gap-3' },
            div({ class: 'form-check form-switch' },
                input({
                    class: 'form-check-input', type: 'checkbox', id: 'push-enabled',
                    checked: () => enabled.val === '1',
                    onchange: event => enabled.val = event.target.checked ? '1' : '0',
                }),
                label({ class: 'form-check-label', for: 'push-enabled' }, '启用推送（NapCat，OneBot v11 HTTP）'),
            ),
            div({ class: 'input-group' },
                span({ class: 'input-group-text' }, 'HTTP 地址'),
                input({
                    class: 'form-control', placeholder: 'http://127.0.0.1:3000',
                    value: url, oninput: event => url.val = event.target.value,
                }),
            ),
            div({ class: 'input-group' },
                span({ class: 'input-group-text' }, 'Token'),
                input({
                    class: 'form-control', type: 'password', autocomplete: 'off',
                    placeholder: 'NapCat 未设置鉴权时留空',
                    value: token, oninput: event => token.val = event.target.value,
                }),
            ),
            div({ class: 'input-group' },
                span({ class: 'input-group-text' }, '推送目标'),
                select({
                    class: 'form-select', style: 'max-width: 9rem;',
                    onchange: event => targetType.val = event.target.value,
                },
                    option({ value: 'private', selected: () => currentTargetType() === 'private' }, '私聊'),
                    option({ value: 'group', selected: () => currentTargetType() === 'group' }, '群聊'),
                ),
                input({
                    class: 'form-control', placeholder: () => currentTargetType() === 'group' ? '群号' : 'QQ 号',
                    value: targetID, oninput: event => targetID.val = event.target.value,
                }),
            ),
            div({ class: 'hstack gap-3' },
                button({ class: 'btn btn-success', onclick: save, disabled: saving }, '保存推送配置'),
                button({ class: 'btn btn-outline-primary', onclick: test, disabled: testing }, '发送测试消息'),
            ),
            div({ class: 'form-text' }, '任务完成、任务失败和订阅新增稿件时会推送到目标 QQ 或群聊；推送失败只记录日志，不影响下载。'),
        ),
    )
}
