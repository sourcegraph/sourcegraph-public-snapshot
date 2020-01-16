import * as GQL from '../../../../../shared/src/graphql/schema'
import React, { FunctionComponent, useCallback, useEffect, useState } from 'react'
import { eventLogger } from '../../../tracking/eventLogger'
import { FilteredConnection, FilteredConnectionQueryArgs } from '../../../components/FilteredConnection'
import { Link } from '../../../../../shared/src/components/Link'
import { PageTitle } from '../../../components/PageTitle'
import { RouteComponentProps } from 'react-router'
import { Timestamp } from '../../../components/time/Timestamp'
import { Collapsible } from '../../../components/Collapsible'
import { fetchLsifUploads } from './backend'
import { Toggle } from '../../../../../shared/src/components/Toggle'

interface HideIncompleteLSIFUploadsToggleProps {
    onlyCompleted: boolean
    onToggle: (enabled: boolean) => void
}

const HideIncompleteLSIFUploadsToggle: FunctionComponent<HideIncompleteLSIFUploadsToggleProps> = ({
    onlyCompleted,
    onToggle,
}) => (
    <div className="lsif-uploads-filter-toggle">
        <label className="radio-buttons__item lsif-uploads-filter-toggle-label" title="Hide incomplete uploads">
            <Toggle value={onlyCompleted} onToggle={onToggle} title="Hide incomplete uploads" />

            <small>
                <div className="radio-buttons__label">Hide incomplete uploads</div>
            </small>
        </label>
    </div>
)

const LsifUploadNode: FunctionComponent<{ node: GQL.ILSIFUpload }> = ({ node }) => (
    <div className="w-100 list-group-item py-2 lsif-data__main">
        <div className="lsif-data__meta">
            <div className="lsif-data__meta-root">
                <code className="e2e-upload-commit">
                    {node.projectRoot?.commit.abbreviatedOID || node.inputCommit.substring(0, 7)}
                </code>
                <span className="ml-2 e2e-upload-root">
                    {node.projectRoot ? (
                        <Link to={node.projectRoot.url}>
                            <strong>{node.projectRoot.path || '/'}</strong>
                        </Link>
                    ) : (
                        node.inputRoot || '/'
                    )}
                </span>
                <span className="ml-2">
                    -
                    <span className="ml-2">
                        <Link to={`./code-intelligence/lsif-uploads/${node.id}`}>
                            {node.state === GQL.LSIFUploadState.PROCESSING ? (
                                <span>Processing</span>
                            ) : node.state === GQL.LSIFUploadState.COMPLETED ? (
                                <span className="text-success">Processed</span>
                            ) : node.state === GQL.LSIFUploadState.ERRORED ? (
                                <span className="text-danger">Failed to process</span>
                            ) : (
                                <span>Waiting to process</span>
                            )}
                        </Link>
                    </span>
                </span>
            </div>
        </div>

        <small className="text-muted lsif-data__meta-timestamp">
            <Timestamp noAbout={true} date={node.finishedAt || node.startedAt || node.uploadedAt} />
        </small>
    </div>
)

interface Props extends RouteComponentProps<{}> {
    repo: GQL.IRepository
}

/**
 * The repository settings code intelligence page.
 */
export const RepoSettingsCodeIntelligencePage: FunctionComponent<Props> = ({ repo, ...props }) => {
    useEffect(() => eventLogger.logViewEvent('RepoSettingsCodeIntelligence'), [])

    // State used in the toggle component shows or hides incomplete uploads
    const [onlyCompleted, setState] = useState(true)

    const queryLatestUploads = useCallback(
        (args: FilteredConnectionQueryArgs) =>
            fetchLsifUploads({ repository: repo.id, isLatestForRepo: true, ...args }),
        [repo.id]
    )

    const queryUploads = useCallback(
        (args: FilteredConnectionQueryArgs) =>
            fetchLsifUploads({
                repository: repo.id,
                ...(onlyCompleted ? { state: GQL.LSIFUploadState.COMPLETED } : {}),
                ...args,
            }),
        [repo.id, onlyCompleted]
    )

    return (
        <div className="repo-settings-code-intelligence-page">
            <PageTitle title="Code intelligence" />
            <h2>Code intelligence</h2>
            <p>
                Enable precise code intelligence by{' '}
                <a href="https://docs.sourcegraph.com/user/code_intelligence/lsif">uploading LSIF data</a>.
            </p>

            <div className="mt-4">
                <h3>Current LSIF uploads</h3>
                <p>
                    These uploads provide code intelligence for the latest commit on the default branch and are used in
                    cross-repository <em>Find References</em> requests.
                </p>

                <FilteredConnection<{}, { node: GQL.ILSIFUpload }>
                    className="list-group list-group-flush mt-3"
                    noun="upload"
                    pluralNoun="uploads"
                    hideSearch={true}
                    useURLQuery={false}
                    noSummaryIfAllNodesVisible={true}
                    queryConnection={queryLatestUploads}
                    nodeComponent={LsifUploadNode}
                    history={props.history}
                    location={props.location}
                    listClassName="list-group list-group-flush"
                    cursorPaging={true}
                    emptyElement={
                        <small>No uploads are recent enough to be used at the tip of the default branch.</small>
                    }
                />
            </div>

            <div className="mt-4">
                <Collapsible
                    title="All LSIF uploads"
                    defaultExpanded={false}
                    className="repo-settings-code-intelligence-page-collapsible"
                    titleClassName="h5 mb-0"
                >
                    <p>These uploads provide code intelligence for branches and older commits.</p>

                    <FilteredConnection<{}, { node: GQL.ILSIFUpload }>
                        className="list-group list-group-flush mt-3"
                        noun="upload"
                        pluralNoun="uploads"
                        queryConnection={queryUploads}
                        nodeComponent={LsifUploadNode}
                        history={props.history}
                        location={props.location}
                        listClassName="list-group list-group-flush"
                        cursorPaging={true}
                        additionalFilterElement={
                            <HideIncompleteLSIFUploadsToggle onlyCompleted={onlyCompleted} onToggle={setState} />
                        }
                    />
                </Collapsible>
            </div>
        </div>
    )
}
