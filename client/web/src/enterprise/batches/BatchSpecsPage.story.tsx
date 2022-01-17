import { storiesOf } from '@storybook/react'
import { addDays } from 'date-fns'
import React from 'react'
import { of } from 'rxjs'

import { WebStory } from '../../../components/WebStory'

import { queryBatchSpecs as _queryBatchSpecs } from './backend'
import { BatchSpecsPage } from './BatchSpecsPage'
import { NODES, successNode } from './testData'

const { add } = storiesOf('web/batches/settings/specs/BatchSpecsPage', module)
    .addDecorator(story => <div className="p-3 container">{story()}</div>)
    .addParameters({
        chromatic: {
            viewports: [320, 576, 978, 1440],
        },
    })

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

add('List of specs', () => (
    <WebStory>{props => <BatchSpecsPage {...props} queryBatchSpecs={queryBatchSpecs} now={NOW} />}</WebStory>
))

add('No specs', () => (
    <WebStory>{props => <BatchSpecsPage {...props} queryBatchSpecs={queryNoBatchSpecs} now={NOW} />}</WebStory>
))
