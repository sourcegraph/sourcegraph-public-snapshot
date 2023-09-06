import type { DecoratorFn, Meta, Story } from '@storybook/react'
import { addDays } from 'date-fns'
import { of } from 'rxjs'

import { WebStory } from '../../components/WebStory'

import type { queryBatchSpecs as _queryBatchSpecs } from './backend'
import { BatchSpecsPage } from './BatchSpecsPage'
import { NODES, successNode } from './testData'

const decorator: DecoratorFn = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/batches/settings/specs/BatchSpecsPage',
    decorators: [decorator],
    parameters: {
        chromatic: {
            viewports: [320, 576, 978, 1440],
            disableSnapshot: false,
        },
    },
}

export default config

const NOW = () => addDays(new Date(), 1)

const queryBatchSpecs: typeof _queryBatchSpecs = () =>
    of({
        __typename: 'BatchSpecConnection',
        totalCount: 47,
        pageInfo: {
            endCursor: null,
            hasNextPage: false,
        },
        nodes: [...NODES, successNode('pid1'), successNode('pid2'), successNode('pid3')],
    })

const queryNoBatchSpecs: typeof _queryBatchSpecs = () =>
    of({
        __typename: 'BatchSpecConnection',
        totalCount: 0,
        pageInfo: {
            endCursor: null,
            hasNextPage: false,
        },
        nodes: [],
    })

export const ListOfSpecs: Story = () => (
    <WebStory>{props => <BatchSpecsPage {...props} queryBatchSpecs={queryBatchSpecs} now={NOW} />}</WebStory>
)

ListOfSpecs.storyName = 'List of specs'

export const NoSpecs: Story = () => (
    <WebStory>{props => <BatchSpecsPage {...props} queryBatchSpecs={queryNoBatchSpecs} now={NOW} />}</WebStory>
)

NoSpecs.storyName = 'No specs'
