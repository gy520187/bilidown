export type IDType = 'bv' | 'ep' | 'ss' | 'fav'

export type CollectionLookup = {
    kind: 'season' | 'series'
    mid: string
    id: string
}

/**
 * 校验用户输入的待解析的视频链接
 * @param url 待解析的视频链接
 * @returns 如果校验成功，则返回 BV 号，否则返回 `false`
 */
export const checkURL = (url: string): {
    type: IDType
    value: string | number
} => {
    const matchBvid = url.match(/(?:https?:\/\/(?:www\.)?bilibili\.com\/video\/)?(BV1[a-zA-Z0-9]+)/i)
    if (matchBvid) return { type: 'bv', value: matchBvid[1] }

    const matchSeason = url.match(/(?:https?:\/\/(?:www\.)?bilibili\.com\/bangumi\/play\/)?(ep|ss)(\d+)/i)
    if (matchSeason) return { type: matchSeason[1].toLowerCase() as 'ep' | 'ss', value: parseInt(matchSeason[2]) }

    try {
        const _url = new URL(url)
        const mediaId = parseInt(_url.searchParams.get('fid') || '')
        if (_url.hostname == 'space.bilibili.com' && _url.pathname.match(/^\/\d+\/favlist$/) && !isNaN(mediaId)) {
            return { type: 'fav', value: mediaId }
        }
        const mlMatch = url.match(/https?:\/\/(?:www\.)?bilibili\.com\/medialist\/detail\/ml(\d+)/)
        if (mlMatch) return { type: 'fav', value: parseInt(mlMatch[1]) }
        const mlShort = url.match(/https?:\/\/(?:www\.)?bilibili\.com\/list\/ml(\d+)/)
        if (mlShort) return { type: 'fav', value: parseInt(mlShort[1]) }
    } catch { }
    throw new Error('您输入的视频链接格式错误')
}

/** 将秒数转换为 `mm:ss` */
export const secondToTime = (second: number) => {
    return `${Math.floor(second / 60)}:${(second % 60).toString().padStart(2, '0')}`
}

export const parseCollectionURL = (url: string): CollectionLookup | false => {
    try { new URL(url) } catch { return false }
    const _url = new URL(url)
    if (_url.hostname !== 'space.bilibili.com') return false

    const listsMatch = _url.pathname.match(/^\/(\d+)\/lists\/(\d+)/)
    if (listsMatch) {
        const listType = _url.searchParams.get('type')
        if (listType === 'series') {
            return { kind: 'series', mid: listsMatch[1], id: listsMatch[2] }
        }
        return { kind: 'season', mid: listsMatch[1], id: listsMatch[2] }
    }

    const mid = _url.pathname.match(/^\/(\d+)\/channel\/collectiondetail$/)?.[1]
    const seasonId = parseInt(_url.searchParams.get('sid') || '')
    if (mid && !isNaN(seasonId)) {
        return { kind: 'season', mid, id: String(seasonId) }
    }

    const seriesMatch = _url.pathname.match(/^\/(\d+)\/channel\/seriesdetail$/)
    const seriesId = parseInt(_url.searchParams.get('sid') || '')
    if (seriesMatch && !isNaN(seriesId)) {
        return { kind: 'series', mid: seriesMatch[1], id: String(seriesId) }
    }
    return false
}
