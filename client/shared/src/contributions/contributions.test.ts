import { describe, expect, test } from 'vitest'

import { ContributableMenu } from '@sourcegraph/client-api'

import type { ActionItemAction } from '../actions/ActionItem'

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
                        'editor/title': [{ action: 'b', alt: 'c' }, { action: 'a' }],
                    },
                },
                ContributableMenu.EditorTitle
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
