import React, { FunctionComponent, useCallback, useEffect } from 'react'
import { RouteComponentProps } from 'react-router'
import { TelemetryProps } from '../../../../../shared/src/telemetry/telemetryService'
import {
    FilteredConnection,
    FilteredConnectionFilter,
    FilteredConnectionQueryArguments,
} from '../../../components/FilteredConnection'
import { PageTitle } from '../../../components/PageTitle'
import { LsifUploadFields, LSIFUploadState, SettingsAreaRepositoryFields } from '../../../graphql-operations'
import { fetchLsifUploads as defaultFetchLsifUploads } from './backend'
import { CodeIntelUploadNode, CodeIntelUploadNodeProps } from './CodeIntelUploadNode'

export interface CodeIntelUploadsPageProps extends RouteComponentProps<{}>, TelemetryProps {
    repo?: SettingsAreaRepositoryFields
    fetchLsifUploads?: typeof defaultFetchLsifUploads
    now?: () => Date
}

const filters: FilteredConnectionFilter[] = [
    {
        id: 'filter',
        type: 'radio',
        label: 'Filter',
        values: [
            {
                label: 'All',
                value: 'all',
                tooltip: 'Show all uploads',
                args: {},
            },
            {
                label: 'Current',
                value: 'current',
                tooltip: 'Show current uploads only',
                args: { isLatestForRepo: true },
            },
            {
                label: 'Completed',
                value: 'completed',
                tooltip: 'Show completed uploads only',
                args: { state: LSIFUploadState.COMPLETED },
            },
            {
                label: 'Errored',
                value: 'errored',
                tooltip: 'Show errored uploads only',
                args: { state: LSIFUploadState.ERRORED },
            },
            {
                label: 'Queued',
                value: 'queued',
                tooltip: 'Show queued uploads only',
                args: { state: LSIFUploadState.QUEUED },
            },
        ],
    },
]

export const CodeIntelUploadsPage: FunctionComponent<CodeIntelUploadsPageProps> = ({
    repo,
    fetchLsifUploads = defaultFetchLsifUploads,
    now,
    telemetryService,
    ...props
}) => {
    useEffect(() => telemetryService.logViewEvent('CodeIntelUploads'), [telemetryService])

    const queryUploads = useCallback(
        (args: FilteredConnectionQueryArguments) => fetchLsifUploads({ repository: repo?.id, ...args }),
        [repo?.id, fetchLsifUploads]
    )

    return (
        <div className="code-intel-uploads web-content">
            <PageTitle title="Precise code intelligence uploads" />
            <h2>Precise code intelligence uploads</h2>
            <p>
                Enable precise code intelligence by{' '}
                <a
                    href="https://docs.sourcegraph.com/code_intelligence/precise_code_intelligence"
                    target="_blank"
                    rel="noreferrer noopener"
                >
                    uploading LSIF data
                </a>
                .
            </p>

            <p>
                Current uploads provide code intelligence for the latest commit on the default branch and are used in
                cross-repository <em>Find References</em> requests. Non-current uploads may still provide code
                intelligence for historic and branch commits.
            </p>

            <div className="list-group position-relative">
                <FilteredConnection<LsifUploadFields, Omit<CodeIntelUploadNodeProps, 'node'>>
                    className="mt-2"
                    listComponent="div"
                    listClassName="codeintel-uploads__grid mb-3"
                    noun="upload"
                    pluralNoun="uploads"
                    nodeComponent={CodeIntelUploadNode}
                    nodeComponentProps={{ now }}
                    queryConnection={queryUploads}
                    history={props.history}
                    location={props.location}
                    cursorPaging={true}
                    filters={filters}
                    defaultFilter="current"
                />
            </div>
        </div>
    )
}
