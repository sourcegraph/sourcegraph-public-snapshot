import React, { useCallback } from 'react'
import { RouteComponentProps } from 'react-router'

import {
    FilteredConnection,
    FilteredConnectionQueryArguments,
} from '@sourcegraph/web/src/components/FilteredConnection'
import { Container, PageHeader } from '@sourcegraph/wildcard'

import { PageTitle } from '../../../components/PageTitle'
import { BatchSpecListFields } from '../../../graphql-operations'

import { queryBatchSpecs as _queryBatchSpecs } from './backend'
import { BatchSpecExecutionNode, BatchSpecExecutionNodeProps } from './BatchSpecExecutionNode'
import styles from './BatchSpecExecutionsPage.module.scss'

export interface BatchSpecExecutionsPageProps extends Pick<RouteComponentProps, 'history' | 'location'> {
    /** For testing purposes only. */
    queryBatchSpecs?: typeof _queryBatchSpecs
    /** For testing purposes only. Sets the current date */
    now?: () => Date
}

export const BatchSpecExecutionsPage: React.FunctionComponent<BatchSpecExecutionsPageProps> = ({
    history,
    location,
    queryBatchSpecs = _queryBatchSpecs,
    now,
}) => {
    const query = useCallback(
        (args: FilteredConnectionQueryArguments) => {
            const passedArguments = {
                first: args.first ?? null,
                after: args.after ?? null,
            }
            return queryBatchSpecs(passedArguments)
        },
        [queryBatchSpecs]
    )

    return (
        <>
            <PageTitle title="Batch spec executions" />
            <PageHeader headingElement="h2" path={[{ text: 'Batch spec executions' }]} className="mb-3" />
            <Container>
                <FilteredConnection<BatchSpecListFields, Omit<BatchSpecExecutionNodeProps, 'node'>>
                    history={history}
                    location={location}
                    nodeComponent={BatchSpecExecutionNode}
                    nodeComponentProps={{ now }}
                    queryConnection={query}
                    hideSearch={true}
                    defaultFirst={20}
                    noun="execution"
                    pluralNoun="executions"
                    listClassName={styles.executionsGrid}
                    listComponent="div"
                    className="filtered-connection__centered-summary"
                    headComponent={ExecutionsHeader}
                    cursorPaging={true}
                    noSummaryIfAllNodesVisible={true}
                    // TODO: This list is just for admins and is not public yet but we
                    // should think about what to show a new Sourcegraph admin when this
                    // list is empty.
                    emptyElement={<>Nobody has executed a batch change yet!</>}
                />
            </Container>
        </>
    )
}

const ExecutionsHeader: React.FunctionComponent<{}> = () => (
    <>
        <span className="d-none d-md-block" />
        <h5 className="p-2 d-none d-md-block text-uppercase text-center text-nowrap">State</h5>
        <h5 className="p-2 d-none d-md-block text-uppercase text-nowrap">Batch spec</h5>
        <h5 className="d-none d-md-block text-uppercase text-center text-nowrap">Execution time</h5>
    </>
)
