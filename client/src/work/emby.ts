const unsafe = /[\\/:*?"<>|\n]/g

export const filterFileName = (name: string): string => name.replace(unsafe, '')

export const sanitizeRelDir = (rel: string): string => {
    return rel.replace(/\\/g, '/').split('/').map(p => filterFileName(p.trim())).filter(p => p && p !== '.' && p !== '..').join('/')
}

const chineseDigits: Record<string, number> = {
    '零': 0, '〇': 0, '一': 1, '二': 2, '两': 2, '三': 3, '四': 4,
    '五': 5, '六': 6, '七': 7, '八': 8, '九': 9, '十': 10,
}

const chineseNumeral = (s: string): number => {
    const t = s.trim()
    if (!t) return 0
    const asNum = Number(t)
    if (Number.isInteger(asNum)) return asNum
    let n = 0
    for (const ch of t) {
        if (/\s/.test(ch)) continue
        const d = chineseDigits[ch]
        if (d === undefined) return 0
        if (d === 10) n = n === 0 ? 10 : n * 10
        else n = n >= 10 ? n + d : d
    }
    return n
}

export const parseSeasonNumber = (seasonTitle: string, extra = false): number => {
    if (extra) return 0
    const s = (seasonTitle || '').trim()
    if (!s) return 1
    const start = s.indexOf('第')
    const end = s.indexOf('季')
    if (start >= 0 && end > start) {
        const n = chineseNumeral(s.slice(start + 1, end))
        if (n > 0) return n
    }
    const m = s.match(/(\d+)/)
    if (m) return Number(m[1])
    return 1
}

export const parseEpisodeNumber = (episodeTitle: string, index: number): number => {
    const s = (episodeTitle || '').trim()
    if (/^\d+$/.test(s)) return Number(s)
    const m = s.match(/(\d+)/)
    if (m) return Number(m[1])
    return index >= 0 ? index + 1 : 1
}

export const pad2 = (n: number): string => String(n).padStart(2, '0')

export type EmbyLayout = {
    show: string
    season: number
    episode: number
    epTitle: string
    relDir: string
    fileStem: string
}

export type WebDLMeta = {
    originName?: string
    evaluate?: string
    year?: number
    videoTag?: string
    videoCodec?: string
    audioCodec?: string
}

export const compactZh = (s: string): string => filterFileName((s || '').trim()).replace(/\s+/g, '')

export const extractEnglish = (...values: string[]): string => {
    for (const s of values) {
        const en = englishFrom(s)
        if (en) return en
    }
    return ''
}

export const dottedEnglish = (s: string): string => {
    const t = (s || '').trim().replace(/[×]/g, 'x').replace(/[·]/g, '.').replace(/[—–\-_\/]/g, ' ')
    let out = ''
    let lastDot = false
    for (const ch of t) {
        if (/[A-Za-z0-9]/.test(ch)) {
            out += ch
            lastDot = false
        } else if (/\s|\./.test(ch)) {
            if (out && !lastDot) {
                out += '.'
                lastDot = true
            }
        }
    }
    return out.replace(/^\.+|\.+$/g, '')
}

const englishFrom = (s: string): string => {
    const t = (s || '').trim()
    if (!t) return ''
    if (isMostlyASCII(t)) return t
    const paren = t.match(/[\(（]([^\)）]{2,})[\)）]/)
    if (paren && isMostlyASCII(paren[1])) return paren[1]
    const runs = t.match(/[A-Za-z][A-Za-z0-9][A-Za-z0-9 .:_'\-]{2,}/g) || []
    let best = ''
    for (const m of runs) {
        if (m.length > best.length && isMostlyASCII(m)) best = m
    }
    return best.trim()
}

const isMostlyASCII = (s: string): boolean => {
    let letters = 0
    let ascii = 0
    for (const ch of s) {
        const code = ch.charCodeAt(0)
        const isAsciiLetter = (code >= 65 && code <= 90) || (code >= 97 && code <= 122)
        const isLetter = isAsciiLetter || code > 127
        if (!isLetter) continue
        letters++
        if (isAsciiLetter) ascii++
    }
    return letters >= 2 && ascii * 2 >= letters
}

export const videoTag = (quality: number, width: number, height: number): string => {
    const h = width > 0 && width < height ? width : height
    if (h >= 2160) return '2160p'
    if (h >= 1440) return '1440p'
    if (h >= 1080) return '1080p'
    if (h >= 720) return '720p'
    if (h >= 480) return '480p'
    if (h >= 360) return '360p'
    if (h > 0) return '240p'
    if (quality === 127) return '4320p'
    if (quality === 120 || quality === 125 || quality === 126) return '2160p'
    if (quality === 112 || quality === 116 || quality === 80) return '1080p'
    if (quality === 74 || quality === 64) return '720p'
    if (quality === 32) return '480p'
    if (quality === 16) return '360p'
    if (quality === 6) return '240p'
    return '1080p'
}

export const codecTag = (codecid: number, codecs = ''): string => {
    if (codecid === 12) return 'H265'
    if (codecid === 13) return 'AV1'
    if (codecid === 7) return 'H264'
    const c = codecs.toLowerCase()
    if (c.includes('hev') || c.includes('hvc') || c.includes('hevc')) return 'H265'
    if (c.includes('av01') || c.includes('av1')) return 'AV1'
    return 'H264'
}

export const audioTag = (preferHiRes: boolean, codecs = '', flac = false): string => {
    const c = codecs.toLowerCase()
    if (flac || (preferHiRes && c.includes('flac'))) return 'FLAC'
    if (c.includes('ec-3') || c.includes('eac3') || c.includes('e-ac-3')) return 'EAC3'
    return 'AAC'
}

export const webDLFileStem = (show: string, meta: WebDLMeta, season: number, episode: number): string => {
    const zh = compactZh(show) || '未命名番剧'
    const en = dottedEnglish(extractEnglish(meta.originName || '', show, meta.evaluate || ''))
    const parts = [zh]
    if (en && en.toLowerCase() !== zh.toLowerCase()) parts.push(en)
    parts.push(`S${pad2(season)}E${pad2(episode)}`)
    if (meta.year && meta.year > 0) parts.push(String(meta.year))
    parts.push(meta.videoTag || '1080p', 'WEB-DL', meta.videoCodec || 'H265', meta.audioCodec || 'AAC', 'bili')
    return filterFileName(parts.join('.'))
}

export const bangumiLayout = (opts: {
    showTitle: string
    seasonTitle: string
    episodeTitle: string
    longTitle: string
    extra?: boolean
    index: number
    meta?: WebDLMeta
}): EmbyLayout => {
    const show = filterFileName((opts.showTitle || '').trim() || '未命名番剧')
    const season = parseSeasonNumber(opts.seasonTitle, !!opts.extra)
    const episode = parseEpisodeNumber(opts.episodeTitle, opts.index)
    let epTitle = (opts.longTitle || '').trim() || (opts.episodeTitle || '').trim()
    if (/^\d+$/.test(epTitle)) epTitle = ''
    epTitle = filterFileName(epTitle)
    return {
        show,
        season,
        episode,
        epTitle,
        relDir: sanitizeRelDir(`${show}/Season ${pad2(season)}`),
        fileStem: webDLFileStem(show, opts.meta || {}, season, episode),
    }
}

export const splitOutRelDir = (title: string, relDir = ''): { title: string, relDir: string } => {
    let rel = sanitizeRelDir(relDir)
    let name = (title || '').replace(/\\/g, '/').trim()
    if (!rel) {
        const i = name.lastIndexOf('/')
        if (i >= 0) {
            rel = sanitizeRelDir(name.slice(0, i))
            name = name.slice(i + 1)
        }
    } else if (name.startsWith(rel + '/')) {
        name = name.slice(rel.length + 1)
    }
    const i = name.lastIndexOf('/')
    if (i >= 0) name = name.slice(i + 1)
    return { title: filterFileName(name), relDir: rel }
}

export const taskFileName = (task: { title: string, id: number, downloadType?: string, relDir?: string }): string => {
    const ext = task.downloadType === 'audio' ? '.m4a' : '.mp4'
    const split = splitOutRelDir(task.title, task.relDir)
    if (split.relDir) {
        return `${split.relDir}/${split.title}${ext}`
    }
    return `${split.title} ${btoa(task.id.toString()).replace(/=/g, '')}${ext}`
}
