import { storiesOf } from '@storybook/react'
import { addDays } from 'date-fns'
import React from 'react'
import { of } from 'rxjs'

import { EnterpriseWebStory } from '../../components/EnterpriseWebStory'

import { queryBatchSpecs as _queryBatchSpecs } from './backend'
import { BatchSpecExecutionsPage } from './BatchSpecExecutionsPage'
import { NODES, successNode } from './testData'

const { add } = storiesOf('web/batches/settings/executions/BatchSpecExecutionsPage', module)
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

const queryNoBatchSpecExecutions: typeof _queryBatchSpecs = () =>
    of({
        __typename: 'BatchSpecConnection',
        totalCount: 0,
        pageInfo: {
            endCursor: null,
            hasNextPage: false,
        },
        nodes: [],
    })

add('List of executions', () => (
    <EnterpriseWebStory>
        {props => <BatchSpecExecutionsPage {...props} queryBatchSpecs={queryBatchSpecs} now={NOW} />}
    </EnterpriseWebStory>
))

add('No executions', () => (
    <EnterpriseWebStory>
        {props => <BatchSpecExecutionsPage {...props} queryBatchSpecs={queryNoBatchSpecExecutions} now={NOW} />}
    </EnterpriseWebStory>
))
