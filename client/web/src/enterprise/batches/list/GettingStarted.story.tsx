import { Meta, Story, DecoratorFn } from '@storybook/react'

import { WebStory } from '../../../components/WebStory'

import { GettingStarted } from './GettingStarted'

const decorator: DecoratorFn = story => <div className="p-3 container">{story()}</div>

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
            defaultValue: false,
        },
        canCreateBatchChanges: {
            control: { type: 'boolean' },
            defaultValue: true,
        },
    },
}

export default config

export const Overview: Story = args => (
    <WebStory>
        {() => <GettingStarted isSourcegraphDotCom={args.isSourcegraphDotCom} canCreate={args.canCreateBatchChanges} />}
    </WebStory>
)
