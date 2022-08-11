import { DecoratorFn, Meta, Story } from '@storybook/react'
import { subHours } from 'date-fns'
import { MATCH_ANY_PARAMETERS, WildcardMockLink } from 'wildcard-mock-link'

import { getDocumentNode } from '@sourcegraph/http-client'
import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'

import { WebStory } from '../../components/WebStory'
import { ExecutorFields, ExecutorsVariables } from '../../graphql-operations'

import { EXECUTORS } from './backend'
import { ExecutorsListPage } from './ExecutorsListPage'

const decorator: DecoratorFn = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/enterprise/executors/ExecutorsListPage',
    decorators: [decorator],
}

export default config

const now = new Date()

const emptyMocks = new WildcardMockLink([
    {
        request: {
            query: getDocumentNode(EXECUTORS),
            variables: MATCH_ANY_PARAMETERS,
        },
        result: {
            data: {
                executors: {
                    __typename: 'ExecutorConnection',
                    nodes: [],
                    totalCount: 0,
                    pageInfo: { endCursor: null, hasNextPage: false },
                },
            },
        },
        nMatches: Number.POSITIVE_INFINITY,
    },
])

const mocks = (count: number) => {
    const nodes = new Array(count).fill(null).map(
        (_value, index): ExecutorFields => ({
            __typename: 'Executor',
            id: index.toString(),
            hostname: `host-${index}`,
            queueName: 'batches',
            active: index % 2 === 0,
            os: 'Linux',
            architecture: 'x86_64',
            dockerVersion: '20.10.17',
            executorVersion: '166204_2022-08-09_433a260eb143',
            gitVersion: '2.37.1',
            igniteVersion: 'v0.10.0',
            srcCliVersion: '3.42.3',
            firstSeenAt: subHours(now, 6).toISOString(),
            lastSeenAt: now.toISOString(),
        })
    )

    return new WildcardMockLink([
        {
            request: {
                query: getDocumentNode(EXECUTORS),
                variables: MATCH_ANY_PARAMETERS,
            },
            result: raw => {
                const variables = raw as ExecutorsVariables
                const filtered = nodes.filter(node => {
                    if (variables.query && !node.hostname.includes(variables.query)) {
                        return false
                    }
                    if (variables.active && !node.active) {
                        return false
                    }

                    return true
                })
                const sliced = filtered.slice(0, variables.first ?? 20)

                return {
                    data: {
                        executors: {
                            __typename: 'ExecutorConnection',
                            nodes: sliced,
                            totalCount: filtered.length,
                            pageInfo: { endCursor: null, hasNextPage: sliced.length < filtered.length },
                        },
                    },
                }
            },
            nMatches: Number.POSITIVE_INFINITY,
        },
    ])
}

export const None: Story = () => (
    <WebStory>
        {props => (
            <MockedTestProvider link={emptyMocks}>
                <ExecutorsListPage {...props} />
            </MockedTestProvider>
        )}
    </WebStory>
)

export const Paginated: Story = () => (
    <WebStory>
        {props => (
            <MockedTestProvider link={mocks(50)}>
                <ExecutorsListPage {...props} />
            </MockedTestProvider>
        )}
    </WebStory>
)

export const Unpaginated: Story = () => (
    <WebStory>
        {props => (
            <MockedTestProvider link={mocks(10)}>
                <ExecutorsListPage {...props} />
            </MockedTestProvider>
        )}
    </WebStory>
)
