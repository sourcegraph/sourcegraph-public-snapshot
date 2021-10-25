import { useCallback } from '@storybook/addons'
import { storiesOf } from '@storybook/react'
import React from 'react'
import { of } from 'rxjs'

import { WebStory } from '../../../components/WebStory'

import { BatchChangeListPage } from './BatchChangeListPage'
import { nodes } from './testData'

const { add } = storiesOf('web/batches/BatchChangeListPage', module)
    .addDecorator(story => <div className="p-3 container">{story()}</div>)
    .addParameters({
        chromatic: {
            viewports: [320, 576, 978, 1440],
        },
    })

const queryBatchChanges = () =>
    of({
        batchChanges: {
            totalCount: Object.values(nodes).length,
            nodes: Object.values(nodes),
            pageInfo: { endCursor: null, hasNextPage: false },
        },
        totalCount: Object.values(nodes).length,
    })

const batchChangesNotLicensed = () => of(false)

const batchChangesLicensed = () => of(true)

add('List of batch changes', () => (
    <WebStory>
        {props => (
            <BatchChangeListPage
                {...props}
                headingElement="h1"
                queryBatchChanges={queryBatchChanges}
                areBatchChangesLicensed={batchChangesLicensed}
            />
        )}
    </WebStory>
))

add('Licensing not enforced', () => (
    <WebStory>
        {props => (
            <BatchChangeListPage
                {...props}
                headingElement="h1"
                queryBatchChanges={queryBatchChanges}
                areBatchChangesLicensed={batchChangesNotLicensed}
            />
        )}
    </WebStory>
))

add('No batch changes', () => {
    const queryBatchChanges = useCallback(
        () =>
            of({
                batchChanges: {
                    totalCount: 0,
                    nodes: [],
                    pageInfo: {
                        endCursor: null,
                        hasNextPage: false,
                    },
                },
                totalCount: 0,
            }),
        []
    )
    return (
        <WebStory>
            {props => (
                <BatchChangeListPage
                    {...props}
                    headingElement="h1"
                    queryBatchChanges={queryBatchChanges}
                    areBatchChangesLicensed={batchChangesLicensed}
                />
            )}
        </WebStory>
    )
})

add('All batch changes tab empty', () => {
    const queryBatchChanges = useCallback(
        () =>
            of({
                batchChanges: {
                    totalCount: 0,
                    nodes: [],
                    pageInfo: {
                        endCursor: null,
                        hasNextPage: false,
                    },
                },
                totalCount: 0,
            }),
        []
    )
    return (
        <WebStory>
            {props => (
                <BatchChangeListPage
                    {...props}
                    headingElement="h1"
                    queryBatchChanges={queryBatchChanges}
                    areBatchChangesLicensed={batchChangesLicensed}
                    openTab="batchChanges"
                />
            )}
        </WebStory>
    )
})
