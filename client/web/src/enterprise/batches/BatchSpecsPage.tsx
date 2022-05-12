import React, { useCallback, useMemo } from 'react'

import MapSearchIcon from 'mdi-react/MapSearchIcon'
import { RouteComponentProps } from 'react-router'

import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { Container, PageHeader, Typography } from '@sourcegraph/wildcard'

import { FilteredConnection, FilteredConnectionQueryArguments } from '../../components/FilteredConnection'
import { PageTitle } from '../../components/PageTitle'
import { BatchSpecListFields, Scalars } from '../../graphql-operations'

import {
    queryBatchSpecs as _queryBatchSpecs,
    queryBatchChangeBatchSpecs as _queryBatchChangeBatchSpecs,
} from './backend'
import { BatchSpecNode, BatchSpecNodeProps } from './BatchSpecNode'

import styles from './BatchSpecsPage.module.scss'

export interface BatchSpecsPageProps extends Omit<BatchSpecListProps, 'currentSpecID'> {}

export const BatchSpecsPage: React.FunctionComponent<React.PropsWithChildren<BatchSpecsPageProps>> = props => (
    <>
        <PageTitle title="Batch specs" />
        <PageHeader
            headingElement="h2"
            path={[{ text: 'Batch specs' }]}
            description="All batch specs that currently exist."
            className="mb-3"
        />
        <Container>
            <BatchSpecList {...props} />
        </Container>
    </>
)

export interface BatchChangeBatchSpecListProps extends Omit<BatchSpecListProps, 'queryBatchSpecs'> {
    batchChangeID: Scalars['ID']
    currentSpecID: Scalars['ID']
    queryBatchChangeBatchSpecs?: typeof _queryBatchChangeBatchSpecs
}

export const BatchChangeBatchSpecList: React.FunctionComponent<
    React.PropsWithChildren<BatchChangeBatchSpecListProps>
> = ({
    history,
    location,
    batchChangeID,
    currentSpecID,
    isLightTheme,
    queryBatchChangeBatchSpecs = _queryBatchChangeBatchSpecs,
    now,
}) => {
    const query = useMemo(() => queryBatchChangeBatchSpecs(batchChangeID), [queryBatchChangeBatchSpecs, batchChangeID])

    return (
        <BatchSpecList
            history={history}
            location={location}
            queryBatchSpecs={query}
            isLightTheme={isLightTheme}
            currentSpecID={currentSpecID}
            now={now}
        />
    )
}

export interface BatchSpecListProps extends ThemeProps, Pick<RouteComponentProps, 'history' | 'location'> {
    currentSpecID?: Scalars['ID']
    queryBatchSpecs?: typeof _queryBatchSpecs
    /** For testing purposes only. Sets the current date */
    now?: () => Date
}

export const BatchSpecList: React.FunctionComponent<React.PropsWithChildren<BatchSpecListProps>> = ({
    history,
    location,
    currentSpecID,
    isLightTheme,
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
        <FilteredConnection<BatchSpecListFields, Omit<BatchSpecNodeProps, 'node'>>
            history={history}
            location={location}
            nodeComponent={BatchSpecNode}
            nodeComponentProps={{ currentSpecID, isLightTheme, now }}
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
    )
}

const Header: React.FunctionComponent<React.PropsWithChildren<{}>> = () => (
    <>
        <span className="d-none d-md-block" />
        <Typography.H5 className="p-2 d-none d-md-block text-uppercase text-center text-nowrap">State</Typography.H5>
        <Typography.H5 className="p-2 d-none d-md-block text-uppercase text-nowrap">Batch spec</Typography.H5>
        <Typography.H5 className="d-none d-md-block text-uppercase text-center text-nowrap">
            Execution time
        </Typography.H5>
    </>
)

const EmptyList: React.FunctionComponent<React.PropsWithChildren<{}>> = () => (
    <div className="text-muted text-center mb-3 w-100">
        <MapSearchIcon className="icon" />
        <div className="pt-2">No batch specs have been created so far.</div>
    </div>
)
