import { checkURL, parseCollectionURL, secondToTime } from './url'

const assertEqual = (actual: unknown, expected: unknown, message: string) => {
    const actualText = JSON.stringify(actual)
    const expectedText = JSON.stringify(expected)
    if (actualText !== expectedText) {
        throw new Error(`${message}: expected ${expectedText}, got ${actualText}`)
    }
}

assertEqual(checkURL('BV1LLDCYJEU3'), { type: 'bv', value: 'BV1LLDCYJEU3' }, 'plain bvid')
assertEqual(
    checkURL('https://www.bilibili.com/video/BV1LLDCYJEU3/'),
    { type: 'bv', value: 'BV1LLDCYJEU3' },
    'video url'
)
assertEqual(
    checkURL('https://www.bilibili.com/bangumi/play/ss48831'),
    { type: 'ss', value: 48831 },
    'season url'
)
assertEqual(
    checkURL('https://space.bilibili.com/1176277996/favlist?fid=1234122612'),
    { type: 'fav', value: 1234122612 },
    'favlist url'
)
assertEqual(
    checkURL('https://www.bilibili.com/list/ml123456'),
    { type: 'fav', value: 123456 },
    'short medialist url'
)
assertEqual(
    checkURL('https://www.bilibili.com/medialist/detail/ml123456'),
    { type: 'fav', value: 123456 },
    'medialist detail url'
)

assertEqual(
    parseCollectionURL('https://space.bilibili.com/282565107/channel/collectiondetail?sid=1427135'),
    { kind: 'season', mid: '282565107', id: '1427135' },
    'legacy collection url'
)
assertEqual(
    parseCollectionURL('https://space.bilibili.com/282565107/lists/1427135?type=season'),
    { kind: 'season', mid: '282565107', id: '1427135' },
    'new collection url'
)
assertEqual(
    parseCollectionURL('https://space.bilibili.com/282565107/lists/998877?type=series'),
    { kind: 'series', mid: '282565107', id: '998877' },
    'new series url'
)
assertEqual(
    parseCollectionURL('https://space.bilibili.com/282565107/channel/seriesdetail?sid=998877'),
    { kind: 'series', mid: '282565107', id: '998877' },
    'legacy series url'
)
assertEqual(parseCollectionURL('https://www.bilibili.com/video/BV1LLDCYJEU3/'), false, 'non collection url')
assertEqual(secondToTime(125), '2:05', 'duration format')

console.log('url tests passed')
