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
                        { id: 'a', command: 'a', title: 'ta', description: 'da', telemetryProps: { feature: 'a' } },
                        { id: 'b', command: 'b', title: 'tb', description: 'db', telemetryProps: { feature: 'b' } },
                        { id: 'c', command: 'c', title: 'tc', description: 'dc', telemetryProps: { feature: 'c' } },
                    ],
                    menus: {
                        'editor/title': [{ action: 'b', alt: 'c' }, { action: 'a' }],
                    },
                },
                ContributableMenu.EditorTitle
            )
        ).toEqual([
            {
                action: { id: 'b', command: 'b', title: 'tb', description: 'db', telemetryProps: { feature: 'b' } },
                active: true,
                altAction: { id: 'c', command: 'c', title: 'tc', description: 'dc', telemetryProps: { feature: 'c' } },
                disabledWhen: false,
            },
            {
                action: { id: 'a', command: 'a', title: 'ta', description: 'da', telemetryProps: { feature: 'a' } },
                active: true,
                altAction: undefined,
                disabledWhen: false,
            },
        ] as ActionItemAction[]))
})
