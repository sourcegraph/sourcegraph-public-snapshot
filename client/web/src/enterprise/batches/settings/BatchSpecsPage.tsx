import MapSearchIcon from 'mdi-react/MapSearchIcon'
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
import { BatchSpecNode, BatchSpecNodeProps } from './BatchSpecNode'
import styles from './BatchSpecsPage.module.scss'

export interface BatchSpecsPageProps extends Pick<RouteComponentProps, 'history' | 'location'> {
    /** For testing purposes only. */
    queryBatchSpecs?: typeof _queryBatchSpecs
    /** For testing purposes only. Sets the current date */
    now?: () => Date
}

export const BatchSpecsPage: React.FunctionComponent<BatchSpecsPageProps> = ({
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
            <PageTitle title="Batch specs" />
            <PageHeader
                headingElement="h2"
                path={[{ text: 'Batch specs' }]}
                description="All batch specs that currently exist."
                className="mb-3"
            />
            <Container>
                <FilteredConnection<BatchSpecListFields, Omit<BatchSpecNodeProps, 'node'>>
                    history={history}
                    location={location}
                    nodeComponent={BatchSpecNode}
                    nodeComponentProps={{ now }}
                    queryConnection={query}
                    hideSearch={true}
                    defaultFirst={20}
                    noun="batch spec"
                    pluralNoun="batch specs"
                    listClassName={styles.specsGrid}
                    listComponent="div"
                    withCenteredSummary={true}
                    headComponent={Header}
                    cursorPaging={true}
                    noSummaryIfAllNodesVisible={true}
                    emptyElement={<EmptyList />}
                />
            </Container>
        </>
    )
}

const Header: React.FunctionComponent<{}> = () => (
    <>
        <span className="d-none d-md-block" />
        <h5 className="p-2 d-none d-md-block text-uppercase text-center text-nowrap">State</h5>
        <h5 className="p-2 d-none d-md-block text-uppercase text-nowrap">Batch spec</h5>
        <h5 className="d-none d-md-block text-uppercase text-center text-nowrap">Execution time</h5>
    </>
)

const EmptyList: React.FunctionComponent<{}> = () => (
    <div className="text-muted text-center mb-3 w-100">
        <MapSearchIcon className="icon" />
        <div className="pt-2">No batch specs have been created so far.</div>
    </div>
)
