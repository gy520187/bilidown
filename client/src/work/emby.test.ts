import { bangumiLayout, dottedEnglish, parseEpisodeNumber, parseSeasonNumber, taskFileName, webDLFileStem } from './emby'

const assertEqual = (actual: unknown, expected: unknown, message: string) => {
    const actualText = JSON.stringify(actual)
    const expectedText = JSON.stringify(expected)
    if (actualText !== expectedText) {
        throw new Error(`${message}: expected ${expectedText}, got ${actualText}`)
    }
}

const layout = bangumiLayout({
    showTitle: '间谍过家家',
    seasonTitle: '第一季',
    episodeTitle: '1',
    longTitle: '任务开始',
    index: 0,
    meta: { originName: 'SPY×FAMILY', year: 2022, videoTag: '1080p', videoCodec: 'H265', audioCodec: 'AAC' },
})
assertEqual(layout.relDir, '间谍过家家/Season 01', 'relDir')
assertEqual(layout.fileStem, '间谍过家家.SPYxFAMILY.S01E01.2022.1080p.WEB-DL.H265.AAC.bili', 'fileStem')
assertEqual(dottedEnglish('The Frontier Under The Blood Moon'), 'The.Frontier.Under.The.Blood.Moon', 'dottedEnglish')
assertEqual(
    webDLFileStem('狄仁杰之狼人劫', { originName: 'The Frontier Under The Blood Moon', year: 2026, videoTag: '2160p' }, 1, 1),
    '狄仁杰之狼人劫.The.Frontier.Under.The.Blood.Moon.S01E01.2026.2160p.WEB-DL.H265.AAC.bili',
    'user template'
)
assertEqual(
    webDLFileStem('闪闪的儿科医生', { originName: 'The GLorious Pediatricians', year: 2023, videoTag: '2160p', videoCodec: 'H265', audioCodec: 'AAC' }, 4, 10),
    '闪闪的儿科医生.The.GLorious.Pediatricians.S04E10.2023.2160p.WEB-DL.H265.AAC.bili',
    'glorious template'
)
assertEqual(parseSeasonNumber('第三季'), 3, '第三季')
assertEqual(parseEpisodeNumber('第08话', 0), 8, '第08话')

const extra = bangumiLayout({
    showTitle: '间谍过家家',
    seasonTitle: '第一季',
    episodeTitle: 'PV',
    longTitle: '预告',
    extra: true,
    index: 0,
})
assertEqual(extra.relDir, '间谍过家家/Season 00', 'extra relDir')

assertEqual(
    taskFileName({ title: '间谍过家家.SPYxFAMILY.S01E01.2022.1080p.WEB-DL.H265.AAC.bili', id: 9, relDir: '间谍过家家/Season 01' }),
    '间谍过家家/Season 01/间谍过家家.SPYxFAMILY.S01E01.2022.1080p.WEB-DL.H265.AAC.bili.mp4',
    'taskFileName emby'
)
assertEqual(
    taskFileName({ title: '闪闪的儿科医生/Season 01/闪闪的儿科医生.The.GLorious.Pediatricians.S04E10.2023.2160p.WEB-DL.H265.AAC.bili', id: 3 }),
    '闪闪的儿科医生/Season 01/闪闪的儿科医生.The.GLorious.Pediatricians.S04E10.2023.2160p.WEB-DL.H265.AAC.bili.mp4',
    'taskFileName baked path'
)

console.log('emby tests passed')
