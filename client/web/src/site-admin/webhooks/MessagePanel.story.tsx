import { number } from '@storybook/addon-knobs'
import { storiesOf } from '@storybook/react'

import { WebStory } from '../../components/WebStory'

import { MessagePanel } from './MessagePanel'
import { BODY_JSON, BODY_PLAIN, HEADERS_JSON, HEADERS_PLAIN } from './story/fixtures'

const { add } = storiesOf('web/site-admin/webhooks/MessagePanel', module)
    .addDecorator(story => <div className="p-3 container">{story()}</div>)
    .addParameters({
        chromatic: {
            viewports: [576, 1440],
        },
    })

for (const [name, { headers, body }] of Object.entries({
    JSON: { headers: HEADERS_JSON, body: BODY_JSON },
    plain: { headers: HEADERS_PLAIN, body: BODY_PLAIN },
})) {
    add(`${name} request`, () => (
        <WebStory>
            {() => (
                <MessagePanel
                    message={{
                        headers,
                        body,
                    }}
                    requestOrStatusCode={{
                        method: 'POST',
                        url: '/my/url',
                        version: 'HTTP/1.1',
                    }}
                />
            )}
        </WebStory>
    ))

    add(`${name} response`, () => (
        <WebStory>
            {() => (
                <MessagePanel
                    message={{
                        headers,
                        body,
                    }}
                    requestOrStatusCode={number('status code', 200, {
                        min: 100,
                        max: 599,
                    })}
                />
            )}
        </WebStory>
    ))
}
