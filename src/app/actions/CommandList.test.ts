import assert from 'assert'
import { filterAndRankItems } from '../CommandList'
import { ActionItemProps } from './ActionItem'

describe('filterAndRankItems', () => {
    function actionIDs(items: ActionItemProps[]): string[] {
        return items.map(({ contribution: { id } }) => id)
    }

    it('no query, no recentActions', () =>
        assert.deepStrictEqual(
            actionIDs(
                filterAndRankItems(
                    [{ contribution: { id: 'a', command: 'a' } }, { contribution: { id: 'b', command: 'b' } }],
                    '',
                    null
                )
            ),
            ['a', 'b']
        ))

    it('query, no recentActions', () =>
        assert.deepStrictEqual(
            actionIDs(
                filterAndRankItems(
                    [
                        { contribution: { id: 'a', command: 'a', title: 'a' } },
                        { contribution: { id: 'b1', command: 'b1', title: 'b' } },
                        { contribution: { id: 'b2', command: 'b2', title: '22b' } },
                    ],
                    'b',
                    null
                )
            ),
            ['b1', 'b2']
        ))

    it('no query, recentActions', () =>
        assert.deepStrictEqual(
            actionIDs(
                filterAndRankItems(
                    [{ contribution: { id: 'a', command: 'a' } }, { contribution: { id: 'b', command: 'b' } }],
                    '',
                    ['b']
                )
            ),
            ['b', 'a']
        ))

    it('query, recentActions', () =>
        assert.deepStrictEqual(
            actionIDs(
                filterAndRankItems(
                    [
                        { contribution: { id: 'a', command: 'a', title: 'a' } },
                        { contribution: { id: 'b1', command: 'b1', title: 'b' } },
                        { contribution: { id: 'b2', command: 'b2', title: '2b' } },
                    ],
                    'b',
                    ['b2']
                )
            ),
            ['b2', 'b1']
        ))
})
