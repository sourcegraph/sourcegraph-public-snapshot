import * as GQL from '../../../../../shared/src/graphql/schema'
import React, { FunctionComponent, useCallback, useEffect, useState, useMemo } from 'react'
import { eventLogger } from '../../../tracking/eventLogger'
import { FilteredConnection, FilteredConnectionQueryArgs } from '../../../components/FilteredConnection'
import { Link } from '../../../../../shared/src/components/Link'
import { PageTitle } from '../../../components/PageTitle'
import { RouteComponentProps } from 'react-router'
import { Timestamp } from '../../../components/time/Timestamp'
import { Collapsible } from '../../../components/Collapsible'
import { deleteLsifUpload, fetchLsifUploads } from './backend'
import { Toggle } from '../../../../../shared/src/components/Toggle'
import DeleteIcon from 'mdi-react/DeleteIcon'
import { ErrorLike, isErrorLike } from '../../../../../shared/src/util/errors'
import { ErrorAlert } from '../../../components/alerts'
import { Subject } from 'rxjs'

interface HideIncompleteLSIFUploadsToggleProps {
    onlyCompleted: boolean
    onToggle: (enabled: boolean) => void
}

const HideIncompleteLSIFUploadsToggle: FunctionComponent<HideIncompleteLSIFUploadsToggleProps> = ({
    onlyCompleted,
    onToggle,
}) => (
    <div className="lsif-uploads-filter-toggle">
        <label className="radio-buttons__item lsif-uploads-filter-toggle-label" title="Show only processed uploads">
            <Toggle value={onlyCompleted} onToggle={onToggle} title="Show only processed uploads" />

            <small>
                <div className="radio-buttons__label">Show only processed uploads</div>
            </small>
        </label>
    </div>
)

const LsifUploadNode: FunctionComponent<{ node: GQL.ILSIFUpload; onDelete: () => void }> = ({ node, onDelete }) => {
    const [deletionOrError, setDeletionOrError] = useState<'loading' | 'deleted' | ErrorLike>()

    const deleteUpload = async (): Promise<void> => {
        let description = `commit ${node.inputCommit.substring(0, 7)}`
        if (node.inputRoot) {
            description += ` rooted at ${node.inputRoot}`
        }

        if (!window.confirm(`Delete upload for commit ${description}?`)) {
            return
        }

        setDeletionOrError('loading')

        try {
            await deleteLsifUpload({ id: node.id }).toPromise()
            onDelete()
        } catch (err) {
            setDeletionOrError(err)
        }
    }

    return deletionOrError && isErrorLike(deletionOrError) ? (
        <div className="alert alert-danger">
            <ErrorAlert prefix="Error deleting LSIF upload" error={deletionOrError} />
        </div>
    ) : (
        <div className="w-100 list-group-item py-2 align-items-center lsif-data__main">
            <div className="lsif-data__meta">
                <div className="lsif-data__meta-root">
                    Upload for commit
                    <code className="ml-1 mr-1 e2e-upload-commit">
                        {node.projectRoot ? (
                            <Link to={node.projectRoot.commit.url}>
                                <code>{node.projectRoot.commit.abbreviatedOID}</code>
                            </Link>
                        ) : (
                            node.inputCommit.substring(0, 7)
                        )}
                    </code>
                    indexed by
                    <span className="ml-1 mr-1">{node.inputIndexer}</span>
                    rooted at
                    <span className="ml-1 e2e-upload-root">
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

                <button
                    type="button"
                    className="btn btn-sm btn-danger lsif-data__meta-delete"
                    onClick={deleteUpload}
                    disabled={deletionOrError === 'loading'}
                    data-tooltip="Delete upload"
                >
                    <DeleteIcon className="icon-inline" />
                </button>
            </small>
        </div>
    )
}

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

    // This observable emits values after successful deletion of an upload and
    // force each filter connection to refresh. As these lists can share the
    // same underlying entity, we need to refresh both at once on deletion of
    // any upload.
    const onDeleteSubject = useMemo(() => new Subject<void>(), [])
    const onDeleteCallback = useMemo(() => onDeleteSubject.next.bind(onDeleteSubject), [onDeleteSubject])

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

                <FilteredConnection<GQL.ILSIFUpload, { onDelete: () => void }>
                    className="list-group list-group-flush mt-3"
                    noun="upload"
                    pluralNoun="uploads"
                    hideSearch={true}
                    useURLQuery={false}
                    noSummaryIfAllNodesVisible={true}
                    queryConnection={queryLatestUploads}
                    nodeComponent={LsifUploadNode}
                    nodeComponentProps={{ onDelete: onDeleteCallback }}
                    updates={onDeleteSubject}
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

                    <FilteredConnection<GQL.ILSIFUpload, { onDelete: () => void }>
                        className="list-group list-group-flush mt-3"
                        noun="upload"
                        pluralNoun="uploads"
                        queryConnection={queryUploads}
                        nodeComponent={LsifUploadNode}
                        nodeComponentProps={{ onDelete: onDeleteCallback }}
                        updates={onDeleteSubject}
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
