import type { Decorator, Meta, StoryFn } from '@storybook/react'

import { WebStory } from '../../components/WebStory'

import { StatusCode } from './StatusCode'

const decorator: Decorator = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/site-admin/webhooks/StatusCode',
    parameters: {},
    decorators: [decorator],
}

export default config

export const Success: StoryFn = args => <WebStory>{() => <StatusCode code={args.code} />}</WebStory>
Success.argTypes = {
    code: {
        control: { type: 'number', min: 100, max: 399 },
    },
}
Success.args = {
    code: 204,
}

export const Failure: StoryFn = args => <WebStory>{() => <StatusCode code={args.code} />}</WebStory>
Failure.argTypes = {
    code: {
        control: { type: 'number', min: 400, max: 599 },
    },
}
Failure.args = {
    code: 418,
}
