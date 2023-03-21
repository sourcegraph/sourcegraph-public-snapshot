import { DecoratorFn, Meta, Story } from '@storybook/react'

import { WebStory } from '../../../components/WebStory'

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

export const Success: Story = args => <WebStory>{() => <StatusCode code={args.code} />}</WebStory>
Success.argTypes = {
    code: {
        control: { type: 'number', min: 100, max: 399 },
        defaultValue: 204,
    },
}

export const Failure: Story = args => <WebStory>{() => <StatusCode code={args.code} />}</WebStory>
Failure.argTypes = {
    code: {
        control: { type: 'number', min: 400, max: 599 },
        defaultValue: 418,
    },
}
