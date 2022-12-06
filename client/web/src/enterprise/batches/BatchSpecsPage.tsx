import React, { useCallback, useMemo } from 'react'

import { mdiMapSearch } from '@mdi/js'
import classNames from 'classnames'
import { RouteComponentProps } from 'react-router'

import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { Container, PageHeader, H3, H5, Icon } from '@sourcegraph/wildcard'

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
                includeLocallyExecutedSpecs: false,
                excludeEmptySpecs: true,
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
            listClassName={classNames(styles.specsGrid, 'test-batches-executions')}
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
        <H5 as={H3} aria-hidden={true} className="p-2 d-none d-md-block text-uppercase text-center text-nowrap">
            State
        </H5>
        <H5 as={H3} aria-hidden={true} className="p-2 d-none d-md-block text-uppercase text-nowrap">
            Batch spec
        </H5>
        <H5 as={H3} aria-hidden={true} className="d-none d-md-block text-uppercase text-center text-nowrap">
            Execution time
        </H5>
    </>
)

const EmptyList: React.FunctionComponent<React.PropsWithChildren<{}>> = () => (
    <div className="text-muted text-center mb-3 w-100">
        <Icon className="icon" svgPath={mdiMapSearch} inline={false} aria-hidden={true} />
        <div className="pt-2">No batch specs have been created so far.</div>
    </div>
)
