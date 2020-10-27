import React, { FunctionComponent, useCallback, useEffect } from 'react'
import { RouteComponentProps } from 'react-router'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { TelemetryProps } from '../../../../../shared/src/telemetry/telemetryService'
import {
    FilteredConnection,
    FilteredConnectionFilter,
    FilteredConnectionQueryArguments,
} from '../../../components/FilteredConnection'
import { LsifIndexFields, LSIFIndexState } from '../../../graphql-operations'
import { fetchLsifIndexes as defaultFetchLsifIndexes } from './backend'
import { CodeIntelIndexNode, CodeIntelIndexNodeProps } from './CodeIntelIndexNode'
import { CodeIntelIndexPageTitle } from './CodeIntelIndexPageTitle'

export interface CodeIntelIndexesPageProps extends RouteComponentProps<{}>, TelemetryProps {
    repo?: GQL.IRepository
    fetchLsifIndexes?: typeof defaultFetchLsifIndexes
    now?: () => Date
}

const filters: FilteredConnectionFilter[] = [
    {
        label: 'All',
        id: 'all',
        tooltip: 'Show all uploads',
        args: {},
    },
    {
        label: 'Completed',
        id: 'completed',
        tooltip: 'Show completed indexes only',
        args: { state: LSIFIndexState.COMPLETED },
    },
    {
        label: 'Errored',
        id: 'errored',
        tooltip: 'Show errored indexes only',
        args: { state: LSIFIndexState.ERRORED },
    },
    {
        label: 'Queued',
        id: 'queued',
        tooltip: 'Show queued indexes only',
        args: { state: LSIFIndexState.QUEUED },
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
        <div className="code-intel-indexes">
            <CodeIntelIndexPageTitle />

            <div className="list-group position-relative">
                <FilteredConnection<LsifIndexFields, Omit<CodeIntelIndexNodeProps, 'node'>>
                    className="mt-2"
                    listComponent="div"
                    listClassName="codeintel-uploads__grid mb-3"
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
