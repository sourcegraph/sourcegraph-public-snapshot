import { Meta, Story, DecoratorFn } from '@storybook/react'

import { WebStory } from '../../../components/WebStory'

import { DotcomGettingStartedPage } from './DotcomGettingStartedPage'

const decorator: DecoratorFn = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/batches/DotcomGettingStartedPage',
    decorators: [decorator],
    parameters: {
        chromatic: {
            disableSnapshot: false,
        },
    },
}

export default config

export const Overview: Story = () => <WebStory>{() => <DotcomGettingStartedPage />}</WebStory>
