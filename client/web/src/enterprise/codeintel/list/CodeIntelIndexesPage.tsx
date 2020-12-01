import React, { FunctionComponent, useCallback, useEffect } from 'react'
import { RouteComponentProps } from 'react-router'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { TelemetryProps } from '../../../../../shared/src/telemetry/telemetryService'
import {
    FilteredConnection,
    FilteredConnectionFilter,
    FilteredConnectionQueryArguments,
} from '../../../components/FilteredConnection'
import { PageTitle } from '../../../components/PageTitle'
import { LsifIndexFields, LSIFIndexState } from '../../../graphql-operations'
import { fetchLsifIndexes as defaultFetchLsifIndexes } from './backend'
import { CodeIntelIndexNode, CodeIntelIndexNodeProps } from './CodeIntelIndexNode'

export interface CodeIntelIndexesPageProps extends RouteComponentProps<{}>, TelemetryProps {
    repo?: GQL.IRepository
    fetchLsifIndexes?: typeof defaultFetchLsifIndexes
    now?: () => Date
}

const filters: FilteredConnectionFilter[] = [
    {
        id: 'filters',
        label: 'Filters',
        type: 'radio',
        values: [
            {
                label: 'All',
                value: 'all',
                tooltip: 'Show all indexes',
                args: {},
            },
            {
                label: 'Completed',
                value: 'completed',
                tooltip: 'Show completed indexes only',
                args: { state: LSIFIndexState.COMPLETED },
            },
            {
                label: 'Errored',
                value: 'errored',
                tooltip: 'Show errored indexes only',
                args: { state: LSIFIndexState.ERRORED },
            },
            {
                label: 'Queued',
                value: 'queued',
                tooltip: 'Show queued indexes only',
                args: { state: LSIFIndexState.QUEUED },
            },
        ],
    },
]

export const CodeIntelIndexesPage: FunctionComponent<CodeIntelIndexesPageProps> = ({
    repo,
    fetchLsifIndexes = defaultFetchLsifIndexes,
    now,
    telemetryService,
    ...props
}) => {
    useEffect(() => telemetryService.logViewEvent('CodeIntelIndexes'), [telemetryService])

    const queryIndexes = useCallback(
        (args: FilteredConnectionQueryArguments) => fetchLsifIndexes({ repository: repo?.id, ...args }),
        [repo?.id, fetchLsifIndexes]
    )

    return (
        <div className="code-intel-indexes web-content">
            <CodeIntelIndexPageTitle />

            <div className="list-group position-relative">
                <FilteredConnection<LsifIndexFields, Omit<CodeIntelIndexNodeProps, 'node'>>
                    className="mt-2"
                    listComponent="div"
                    listClassName="codeintel-indexes__grid mb-3"
                    noun="index"
                    pluralNoun="indexes"
                    nodeComponent={CodeIntelIndexNode}
                    nodeComponentProps={{ now }}
                    queryConnection={queryIndexes}
                    history={props.history}
                    location={props.location}
                    cursorPaging={true}
                    filters={filters}
                    defaultFilter="All"
                />
            </div>
        </div>
    )
}

const CodeIntelIndexPageTitle: FunctionComponent<{}> = () => (
    <>
        <PageTitle title="Precise code intelligence auto-index records" />
        <h2>Precise code intelligence auto-index records</h2>
        <p>
            Popular Go repositories are indexed automatically via{' '}
            <a href="https://github.com/sourcegraph/lsif-go" target="_blank" rel="noreferrer noopener">
                lsif-go
            </a>{' '}
            on{' '}
            <a href="https://sourcegraph.com" target="_blank" rel="noreferrer noopener">
                Sourcegraph.com
            </a>
            . Enable precise code intelligence for non-Go code by{' '}
            <a
                href="https://docs.sourcegraph.com/code_intelligence/precise_code_intelligence"
                target="_blank"
                rel="noreferrer noopener"
            >
                uploading LSIF data
            </a>
            .
        </p>
    </>
)
