import { number } from '@storybook/addon-knobs'
import { DecoratorFn, Meta, Story } from '@storybook/react'

import { WebStory } from '../../components/WebStory'

import { StatusCode } from './StatusCode'

const decorator: DecoratorFn = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/site-admin/webhooks/StatusCode',
    parameters: {
        chromatic: {
            viewports: [576],
        },
    },
    decorators: [decorator],
}

export default config

export const Success: Story = () => (
    <WebStory>{() => <StatusCode code={number('code', 204, { min: 100, max: 399 })} />}</WebStory>
)

export const Failure: Story = () => (
    <WebStory>{() => <StatusCode code={number('code', 418, { min: 400, max: 599 })} />}</WebStory>
)
