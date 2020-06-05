import * as GQL from '../../../../shared/src/graphql/schema'
import React, { FunctionComponent, useCallback, useEffect, useState, useMemo } from 'react'
import { eventLogger } from '../../tracking/eventLogger'
import {
    FilteredConnection,
    FilteredConnectionQueryArgs,
    FilteredConnectionFilter,
} from '../../components/FilteredConnection'
import { Link } from '../../../../shared/src/components/Link'
import { PageTitle } from '../../components/PageTitle'
import { RouteComponentProps } from 'react-router'
import { Timestamp } from '../../components/time/Timestamp'
import { deleteLsifUpload, fetchLsifUploads } from './backend'
import DeleteIcon from 'mdi-react/DeleteIcon'
import { ErrorLike, isErrorLike } from '../../../../shared/src/util/errors'
import { ErrorAlert } from '../../components/alerts'
import { Subject } from 'rxjs'
import * as H from 'history'

export interface LsifUploadNodeProps {
    node: GQL.ILSIFUpload
    onDelete: () => void
    history: H.History
}

const LsifUploadNode: FunctionComponent<LsifUploadNodeProps> = ({ node, onDelete, history }) => {
    const [deletionOrError, setDeletionOrError] = useState<'loading' | 'deleted' | ErrorLike>()

    const deleteUpload = async (): Promise<void> => {
        let description = `commit ${node.inputCommit.slice(0, 7)}`
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
        } catch (error) {
            setDeletionOrError(error)
        }
    }

    return deletionOrError && isErrorLike(deletionOrError) ? (
        <ErrorAlert prefix="Error deleting LSIF upload" error={deletionOrError} history={history} />
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
                            node.inputCommit.slice(0, 7)
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
                            <Link to={`./uploads/${node.id}`}>
                                {node.state === GQL.LSIFUploadState.UPLOADING ? (
                                    <span>Uploading</span>
                                ) : node.state === GQL.LSIFUploadState.PROCESSING ? (
                                    <span>Processing</span>
                                ) : node.state === GQL.LSIFUploadState.COMPLETED ? (
                                    <span className="text-success">Processed</span>
                                ) : node.state === GQL.LSIFUploadState.ERRORED ? (
                                    <span className="text-danger">Failed to process</span>
                                ) : (
                                    <span>Waiting to process (#{node.placeInQueue} in line)</span>
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
    repo?: GQL.IRepository
}

/**
 * The repository settings code intel uploads page.
 */
export const CodeIntelUploadsPage: FunctionComponent<Props> = ({ repo, ...props }) => {
    useEffect(() => eventLogger.logViewEvent('CodeIntelUploads'), [])

    const filters: FilteredConnectionFilter[] = [
        {
            label: 'Only current',
            id: 'current',
            tooltip: 'Show current uploads only',
            args: { isLatestForRepo: true },
        },
        {
            label: 'Only completed',
            id: 'completed',
            tooltip: 'Show completed uploads only',
            args: { state: GQL.LSIFUploadState.COMPLETED },
        },
        {
            label: 'All',
            id: 'all',
            tooltip: 'Show all uploads',
            args: {},
        },
    ]

    // This observable emits values after successful deletion of an upload and
    // forces the filter connection to refresh.
    const onDeleteSubject = useMemo(() => new Subject<void>(), [])
    const onDeleteCallback = useMemo(() => onDeleteSubject.next.bind(onDeleteSubject), [onDeleteSubject])

    const queryUploads = useCallback(
        (args: FilteredConnectionQueryArgs) => fetchLsifUploads({ repository: repo?.id, ...args }),
        [repo?.id]
    )

    return (
        <div className="repo-settings-code-intelligence-page">
            <PageTitle title="Code intelligence - uploads" />
            <h2>Code intelligence - precise code intel uploads</h2>
            <p>
                Enable precise code intelligence by{' '}
                <a
                    href="https://docs.sourcegraph.com/user/code_intelligence/lsif"
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

            <FilteredConnection<GQL.ILSIFUpload, Omit<LsifUploadNodeProps, 'node'>>
                className="list-group list-group-flush mt-3"
                noun="upload"
                pluralNoun="uploads"
                queryConnection={queryUploads}
                nodeComponent={LsifUploadNode}
                nodeComponentProps={{ onDelete: onDeleteCallback, history: props.history }}
                updates={onDeleteSubject}
                history={props.history}
                location={props.location}
                listClassName="list-group list-group-flush"
                cursorPaging={true}
                filters={filters}
            />
        </div>
    )
}
