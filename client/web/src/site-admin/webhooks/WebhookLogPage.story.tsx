import { storiesOf } from '@storybook/react'
import { addMinutes, formatRFC3339 } from 'date-fns'
import { of } from 'rxjs'

import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'

import { WebStory } from '../../components/WebStory'
import { WebhookLogFields, WebhookLogsVariables } from '../../graphql-operations'

import { queryWebhookLogs, SelectedExternalService } from './backend'
import { BODY_JSON, BODY_PLAIN, buildHeaderMock, HEADERS_JSON, HEADERS_PLAIN } from './story/fixtures'
import { WebhookLogPage } from './WebhookLogPage'

const { add } = storiesOf('web/site-admin/webhooks/WebhookLogPage', module)
    .addDecorator(story => <div className="p-3 container">{story()}</div>)
    .addParameters({
        chromatic: {
            viewports: [320, 576, 978, 1440],
        },
    })

const buildQueryWebhookLogs: (logs: WebhookLogFields[]) => typeof queryWebhookLogs = logs => (
    { first, after }: Pick<WebhookLogsVariables, 'first' | 'after'>,
    externalService: SelectedExternalService,
    onlyErrors: boolean
) => {
    const filtered = logs.filter(log => {
        if (onlyErrors && log.statusCode < 400) {
            return false
        }

        if (externalService === 'unmatched' && log.externalService) {
            return false
        }
        if (
            externalService !== 'all' &&
            externalService !== 'unmatched' &&
            externalService !== log.externalService?.displayName
        ) {
            return false
        }

        return true
    })

    first = first ?? 20
    const afterNumber = after?.length ? +after : 0
    const page = filtered.slice(afterNumber, afterNumber + first)
    const cursor = afterNumber + first

    return of({
        nodes: page,
        pageInfo: {
            hasNextPage: logs.length > cursor,
            endCursor: cursor.toString(),
        },
        totalCount: logs.length,
    })
}

const buildWebhookLogs = (count: number, externalServiceCount: number): WebhookLogFields[] => {
    const logs: WebhookLogFields[] = []
    const time = new Date(2021, 10, 8, 16, 40, 30)

    for (let index = 0; index < count; index++) {
        const externalServiceID = index % (externalServiceCount + 1)
        const statusCode =
            index % 3 === 0
                ? 200 + Math.floor(index / 3)
                : index % 3 === 1
                ? 400 + Math.floor(index / 3)
                : 500 + Math.floor(index / 3)

        logs.push({
            id: index.toString(),
            receivedAt: formatRFC3339(addMinutes(time, index)),
            externalService:
                externalServiceID === externalServiceCount
                    ? null
                    : {
                          displayName: `External service ${externalServiceID}`,
                      },
            statusCode,
            request: {
                headers: HEADERS_JSON,
                body: BODY_JSON,
                method: 'POST',
                url: '/my/url',
                version: 'HTTP/1.1',
            },
            response: {
                headers: HEADERS_PLAIN,
                body: BODY_PLAIN,
            },
        })
    }

    return logs
}

add('no logs', () => (
    <WebStory>
        {props => (
            <MockedTestProvider mocks={buildHeaderMock(2, 2)}>
                <WebhookLogPage {...props} queryWebhookLogs={buildQueryWebhookLogs([])} />
            </MockedTestProvider>
        )}
    </WebStory>
))

add('one page of logs', () => (
    <WebStory>
        {props => (
            <MockedTestProvider mocks={buildHeaderMock(2, 2)}>
                <WebhookLogPage {...props} queryWebhookLogs={buildQueryWebhookLogs(buildWebhookLogs(20, 2))} />
            </MockedTestProvider>
        )}
    </WebStory>
))

add('two pages of logs', () => (
    <WebStory>
        {props => (
            <MockedTestProvider mocks={buildHeaderMock(2, 2)}>
                <WebhookLogPage {...props} queryWebhookLogs={buildQueryWebhookLogs(buildWebhookLogs(40, 2))} />
            </MockedTestProvider>
        )}
    </WebStory>
))
