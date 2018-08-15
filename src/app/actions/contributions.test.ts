import assert from 'assert'
import { ContributableMenu } from 'cxp/module/protocol'
import { ActionItemProps } from './ActionItem'
import { getContributedActionItems } from './contributions'

describe('getContributedActionItems', () => {
    it('gets action items', () =>
        assert.deepStrictEqual(
            getContributedActionItems(
                {
                    actions: [
                        { id: 'a', command: 'a', title: 'ta', description: 'da' },
                        { id: 'b', command: 'b', title: 'tb', description: 'db' },
                        { id: 'c', command: 'c', title: 'tc', description: 'dc' },
                    ],
                    menus: {
                        commandPalette: [{ action: 'a' }, { action: 'b' }],
                        'editor/title': [{ action: 'c' }],
                    },
                },
                ContributableMenu.CommandPalette
            ),
            [
                { contribution: { id: 'a', command: 'a', title: 'ta', description: 'da' } },
                { contribution: { id: 'b', command: 'b', title: 'tb', description: 'db' } },
            ] as ActionItemProps[]
        ))
})
