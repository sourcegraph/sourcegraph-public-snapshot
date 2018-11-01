import { assert } from 'chai'
import { filterAndRankItems } from '../CommandList'
import { ActionItemProps } from './ActionItem'

describe('filterAndRankItems', () => {
    function actionIDs(items: ActionItemProps[]): string[] {
        return items.map(({ action: { id } }) => id)
    }

    it('no query, no recentActions', () =>
        assert.deepEqual(
            actionIDs(
                filterAndRankItems(
                    [{ action: { id: 'a', command: 'a' } }, { action: { id: 'b', command: 'b' } }],
                    '',
                    null
                )
            ),
            ['a', 'b']
        ))

    it('query, no recentActions', () =>
        assert.deepEqual(
            actionIDs(
                filterAndRankItems(
                    [
                        { action: { id: 'a', command: 'a', title: 'a' } },
                        { action: { id: 'b1', command: 'b1', title: 'b' } },
                        { action: { id: 'b2', command: 'b2', title: '22b' } },
                    ],
                    'b',
                    null
                )
            ),
            ['b1', 'b2']
        ))

    it('no query, recentActions', () =>
        assert.deepEqual(
            actionIDs(
                filterAndRankItems([{ action: { id: 'a', command: 'a' } }, { action: { id: 'b', command: 'b' } }], '', [
                    'b',
                ])
            ),
            ['b', 'a']
        ))

    it('query, recentActions', () =>
        assert.deepEqual(
            actionIDs(
                filterAndRankItems(
                    [
                        { action: { id: 'a', command: 'a', title: 'a' } },
                        { action: { id: 'b1', command: 'b1', title: 'b' } },
                        { action: { id: 'b2', command: 'b2', title: '2b' } },
                    ],
                    'b',
                    ['b2']
                )
            ),
            ['b2', 'b1']
        ))
})
