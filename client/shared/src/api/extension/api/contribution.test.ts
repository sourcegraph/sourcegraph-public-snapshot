import { ContributableMenu, Contributions } from '@sourcegraph/client-api'

import { filterContributions, mergeContributions } from './contribution'

describe('mergeContributions()', () => {
    const FIXTURE_CONTRIBUTIONS_1: Contributions = {
        actions: [
            { id: '1.a', command: 'c', title: '1.A' },
            { id: '1.b', command: 'c', title: '1.B' },
        ],
        menus: {
            [ContributableMenu.GlobalNav]: [{ action: '1.a' }, { action: '1.b' }],
        },
    }

    const FIXTURE_CONTRIBUTIONS_2: Contributions = {
        actions: [
            { id: '2.a', command: 'c', title: '2.A' },
            { id: '2.b', command: 'c', title: '2.B' },
        ],
        menus: {
            [ContributableMenu.EditorTitle]: [{ action: '2.a' }, { action: '2.b' }],
        },
    }
    const FIXTURE_CONTRIBUTIONS_MERGED: Contributions = {
        actions: [
            { id: '1.a', command: 'c', title: '1.A' },
            { id: '1.b', command: 'c', title: '1.B' },
            { id: '2.a', command: 'c', title: '2.A' },
            { id: '2.b', command: 'c', title: '2.B' },
        ],
        menus: {
            [ContributableMenu.GlobalNav]: [{ action: '1.a' }, { action: '1.b' }],
            [ContributableMenu.EditorTitle]: [{ action: '2.a' }, { action: '2.b' }],
        },
    }

    test('handles an empty array', () => {
        expect(mergeContributions([])).toEqual({})
    })
    test('handles a single item', () => {
        expect(mergeContributions([FIXTURE_CONTRIBUTIONS_1])).toEqual(FIXTURE_CONTRIBUTIONS_1)
    })
    test('handles multiple items', () => {
        expect(
            mergeContributions([FIXTURE_CONTRIBUTIONS_1, FIXTURE_CONTRIBUTIONS_2, {}, { actions: [] }, { menus: {} }])
        ).toEqual(FIXTURE_CONTRIBUTIONS_MERGED)
    })
})

describe('filterContributions()', () => {
    it('handles empty contributions', () => {
        expect(filterContributions({})).toEqual({})
    })

    it('handles empty menu contributions', () => {
        const expected: Contributions = {
            menus: {},
        }
        expect(filterContributions({ menus: {} })).toEqual(expected)
    })

    it('handles non-empty contributions', () => {
        const expected: Contributions = {
            actions: [
                { id: 'a1', command: 'c' },
                { id: 'a2', command: 'c' },
                { id: 'a3', command: 'c' },
            ],
            menus: {
                [ContributableMenu.GlobalNav]: [{ action: 'a1', when: true }, { action: 'a2' }],
            },
        }
        expect(
            filterContributions({
                actions: [
                    { id: 'a1', command: 'c' },
                    { id: 'a2', command: 'c' },
                    { id: 'a3', command: 'c' },
                ],
                menus: {
                    [ContributableMenu.GlobalNav]: [
                        { action: 'a1', when: true },
                        { action: 'a2' },
                        { action: 'a3', when: false },
                    ],
                },
            })
        ).toEqual(expected)
    })
})
/* eslint-enable no-template-curly-in-string */
