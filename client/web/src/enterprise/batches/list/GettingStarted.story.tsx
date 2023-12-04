import type { Meta, StoryFn, Decorator } from '@storybook/react'

import { WebStory } from '../../../components/WebStory'

import { GettingStarted } from './GettingStarted'

const decorator: Decorator = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/batches/GettingStarted',
    decorators: [decorator],
    parameters: {
        chromatic: {
            disableSnapshot: false,
        },
    },
    argTypes: {
        isSourcegraphDotCom: {
            control: { type: 'boolean' },
        },
        canCreateBatchChanges: {
            control: { type: 'boolean' },
        },
    },
    args: {
        isSourcegraphDotCom: false,
        canCreateBatchChanges: true,
    },
}

export default config

export const Overview: StoryFn = args => (
    <WebStory>
        {() => <GettingStarted isSourcegraphDotCom={args.isSourcegraphDotCom} canCreate={args.canCreateBatchChanges} />}
    </WebStory>
)
