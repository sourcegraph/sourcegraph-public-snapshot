import { ContributableMenu } from '@sourcegraph/client-api'

import { ActionItemAction } from '../actions/ActionItem'

import { getContributedActionItems } from './contributions'

describe('getContributedActionItems', () => {
    test('gets action items', () =>
        expect(
            getContributedActionItems(
                {
                    actions: [
                        { id: 'a', command: 'a', title: 'ta', description: 'da' },
                        { id: 'b', command: 'b', title: 'tb', description: 'db' },
                        { id: 'c', command: 'c', title: 'tc', description: 'dc' },
                    ],
                    menus: {
                        commandPalette: [
                            { action: 'a', group: '2' },
                            { action: 'b', group: '1', alt: 'c' },
                        ],
                        'editor/title': [{ action: 'c' }],
                    },
                },
                ContributableMenu.CommandPalette
            )
        ).toEqual([
            {
                action: { id: 'b', command: 'b', title: 'tb', description: 'db' },
                active: true,
                altAction: { id: 'c', command: 'c', title: 'tc', description: 'dc' },
                disabledWhen: false,
            },
            {
                action: { id: 'a', command: 'a', title: 'ta', description: 'da' },
                active: true,
                altAction: undefined,
                disabledWhen: false,
            },
        ] as ActionItemAction[]))
})
