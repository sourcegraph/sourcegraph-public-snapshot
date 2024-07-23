import type { Decorator, Meta, StoryFn } from '@storybook/react'

import { WebStory } from '../../components/WebStory'

import { MessagePanel } from './MessagePanel'
import { BODY_JSON, BODY_PLAIN, HEADERS_JSON, HEADERS_PLAIN } from './story/fixtures'

const decorator: Decorator = story => <div className="p-3 container">{story()}</div>
const config: Meta = {
    title: 'web/site-admin/webhooks/MessagePanel',
    decorators: [decorator],
    parameters: {},
}

export default config

const messagePanelObject = {
    JSON: { headers: HEADERS_JSON, body: BODY_JSON },
    plain: { headers: HEADERS_PLAIN, body: BODY_PLAIN },
}

export const JSONRequest: StoryFn = () => (
    <WebStory>
        {() => (
            <MessagePanel
                message={{
                    headers: messagePanelObject.JSON.headers,
                    body: messagePanelObject.JSON.body,
                }}
                requestOrStatusCode={{
                    method: 'POST',
                    url: '/my/url',
                    version: 'HTTP/1.1',
                }}
            />
        )}
    </WebStory>
)

JSONRequest.storyName = 'JSON request'

export const JSONResponse: StoryFn = args => (
    <WebStory>
        {() => (
            <MessagePanel
                message={{
                    headers: messagePanelObject.JSON.headers,
                    body: messagePanelObject.JSON.body,
                }}
                requestOrStatusCode={args.requestOrStatusCode}
            />
        )}
    </WebStory>
)
JSONResponse.argTypes = {
    requestOrStatusCode: {
        control: { type: 'number', min: 100, max: 599 },
    },
}
JSONResponse.args = {
    requestOrStatusCode: 200,
}

JSONResponse.storyName = 'JSON response'

export const PlainRequest: StoryFn = () => (
    <WebStory>
        {() => (
            <MessagePanel
                message={{
                    headers: messagePanelObject.plain.headers,
                    body: messagePanelObject.plain.body,
                }}
                requestOrStatusCode={{
                    method: 'POST',
                    url: '/my/url',
                    version: 'HTTP/1.1',
                }}
            />
        )}
    </WebStory>
)

PlainRequest.storyName = 'plain request'

export const PlainResponse: StoryFn = args => (
    <WebStory>
        {() => (
            <MessagePanel
                message={{
                    headers: messagePanelObject.plain.headers,
                    body: messagePanelObject.plain.body,
                }}
                requestOrStatusCode={args.requestOrStatusCode}
            />
        )}
    </WebStory>
)
PlainResponse.argTypes = {
    requestOrStatusCode: {
        control: { type: 'number', min: 100, max: 599 },
    },
}
PlainResponse.args = {
    requestOrStatusCode: 200,
}

PlainResponse.storyName = 'plain response'
