import { ActionItemAction } from '../actions/ActionItem'

import { filterAndRankItems } from './CommandList'

describe('filterAndRankItems', () => {
    function actionIDs(items: ActionItemAction[]): string[] {
        return items.map(({ action: { id } }) => id)
    }

    test('no query, no recentActions', () =>
        expect(
            actionIDs(
                filterAndRankItems(
                    [
                        { action: { id: 'a', command: 'a' }, active: true },
                        { action: { id: 'b', command: 'b' }, active: true },
                    ],
                    '',
                    null
                )
            )
        ).toEqual(['a', 'b']))

    test('query, no recentActions', () =>
        expect(
            actionIDs(
                filterAndRankItems(
                    [
                        { action: { id: 'a', command: 'a', title: 'a' }, active: true },
                        { action: { id: 'b1', command: 'b1', title: 'b' }, active: true },
                        { action: { id: 'b2', command: 'b2', title: '22b' }, active: true },
                    ],
                    'b',
                    null
                )
            )
        ).toEqual(['b1', 'b2']))

    test('no query, recentActions', () =>
        expect(
            actionIDs(
                filterAndRankItems(
                    [
                        { action: { id: 'a', command: 'a' }, active: true },
                        { action: { id: 'b', command: 'b' }, active: true },
                    ],
                    '',
                    ['b']
                )
            )
        ).toEqual(['b', 'a']))

    test('query, recentActions', () =>
        expect(
            actionIDs(
                filterAndRankItems(
                    [
                        { action: { id: 'a', command: 'a', title: 'a' }, active: true },
                        { action: { id: 'b1', command: 'b1', title: 'b' }, active: true },
                        { action: { id: 'b2', command: 'b2', title: '2b' }, active: true },
                    ],
                    'b',
                    ['b2']
                )
            )
        ).toEqual(['b2', 'b1']))
})
