import type { Decorator, Meta, StoryFn } from '@storybook/react'
import classNames from 'classnames'
import { addDays } from 'date-fns'
import { MATCH_ANY_PARAMETERS, WildcardMockLink } from 'wildcard-mock-link'

import { getDocumentNode } from '@sourcegraph/http-client'
import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'

import { WebStory } from '../../components/WebStory'

import { BATCH_SPEC_WORKSPACE_FILE } from './backend'
import { BatchSpecNode } from './BatchSpecNode'
import { NODES, MOCK_HIGHLIGHTED_FILES } from './testData'

import styles from './BatchSpecsPage.module.scss'

const NOW = () => addDays(new Date(), 1)

const decorator: Decorator = story => <div className={classNames(styles.specsGrid, 'p-3 container')}>{story()}</div>

const config: Meta = {
    title: 'web/batches/batch-spec',
    decorators: [decorator],
}

export default config

export const BatchSpecNodeStory: StoryFn = () => {
    const mocks = new WildcardMockLink([
        {
            request: {
                query: getDocumentNode(BATCH_SPEC_WORKSPACE_FILE),
                variables: MATCH_ANY_PARAMETERS,
            },
            result: { data: MOCK_HIGHLIGHTED_FILES },
            nMatches: Number.POSITIVE_INFINITY,
        },
    ])

    return (
        <WebStory>
            {props => (
                <MockedTestProvider link={mocks}>
                    <>
                        {NODES.map(node => (
                            <BatchSpecNode {...props} key={node.id} node={node} now={NOW} />
                        ))}
                    </>
                </MockedTestProvider>
            )}
        </WebStory>
    )
}

BatchSpecNodeStory.storyName = 'BatchSpecNode'
