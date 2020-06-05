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
import { deleteLsifIndex, fetchLsifIndexes } from './backend'
import DeleteIcon from 'mdi-react/DeleteIcon'
import { ErrorLike, isErrorLike } from '../../../../shared/src/util/errors'
import { ErrorAlert } from '../../components/alerts'
import { Subject } from 'rxjs'
import * as H from 'history'

export interface LsifIndexNodeProps {
    node: GQL.ILSIFIndex
    onDelete: () => void
    history: H.History
}

const LsifIndexNode: FunctionComponent<LsifIndexNodeProps> = ({ node, onDelete, history }) => {
    const [deletionOrError, setDeletionOrError] = useState<'loading' | 'deleted' | ErrorLike>()

    const deleteIndex = async (): Promise<void> => {
        const description = `commit ${node.inputCommit.slice(0, 7)}`

        if (!window.confirm(`Delete index for commit ${description}?`)) {
            return
        }

        setDeletionOrError('loading')

        try {
            await deleteLsifIndex({ id: node.id }).toPromise()
            onDelete()
        } catch (error) {
            setDeletionOrError(error)
        }
    }

    return deletionOrError && isErrorLike(deletionOrError) ? (
        <ErrorAlert prefix="Error deleting LSIF index" error={deletionOrError} history={history} />
    ) : (
        <div className="w-100 list-group-item py-2 align-items-center lsif-data__main">
            <div className="lsif-data__meta">
                <div className="lsif-data__meta-root">
                    Index for commit
                    <code className="ml-1 mr-1 e2e-index-commit">
                        {node.projectRoot ? (
                            <Link to={node.projectRoot.commit.url}>
                                <code>{node.projectRoot.commit.abbreviatedOID}</code>
                            </Link>
                        ) : (
                            node.inputCommit.slice(0, 7)
                        )}
                    </code>
                    rooted at
                    <span className="ml-1 e2e-index-root">
                        {node.projectRoot ? (
                            <Link to={node.projectRoot.url}>
                                <strong>{node.projectRoot.path || '/'}</strong>
                            </Link>
                        ) : (
                            '/'
                        )}
                    </span>
                    <span className="ml-2">
                        -
                        <span className="ml-2">
                            <Link to={`./indexes/${node.id}`}>
                                {node.state === GQL.LSIFIndexState.PROCESSING ? (
                                    <span>Processing</span>
                                ) : node.state === GQL.LSIFIndexState.COMPLETED ? (
                                    <span className="text-success">Processed</span>
                                ) : node.state === GQL.LSIFIndexState.ERRORED ? (
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
                <Timestamp noAbout={true} date={node.finishedAt || node.startedAt || node.queuedAt} />

                <button
                    type="button"
                    className="btn btn-sm btn-danger lsif-data__meta-delete"
                    onClick={deleteIndex}
                    disabled={deletionOrError === 'loading'}
                    data-tooltip="Delete index"
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
 * The repository settings code intelligence page.
 */
export const CodeIntelIndexesPage: FunctionComponent<Props> = ({ repo, ...props }) => {
    useEffect(() => eventLogger.logViewEvent('CodeIntelIndexes'), [])

    const filters: FilteredConnectionFilter[] = [
        {
            label: 'Only current',
            id: 'current',
            tooltip: 'Show current indexes only',
            args: { isLatestForRepo: true },
        },
        {
            label: 'Only completed',
            id: 'completed',
            tooltip: 'Show completed indexes only',
            args: { state: GQL.LSIFIndexState.COMPLETED },
        },
        {
            label: 'All',
            id: 'all',
            tooltip: 'Show all indexes',
            args: {},
        },
    ]

    // This observable emits values after successful deletion of an index and
    // forces the filter connection to refresh.
    const onDeleteSubject = useMemo(() => new Subject<void>(), [])
    const onDeleteCallback = useMemo(() => onDeleteSubject.next.bind(onDeleteSubject), [onDeleteSubject])

    const queryIndexes = useCallback(
        (args: FilteredConnectionQueryArgs) => fetchLsifIndexes({ repository: repo?.id, ...args }),
        [repo?.id]
    )

    return (
        <div className="repo-settings-code-intelligence-page">
            <PageTitle title="Code intelligence - auto-indexing" />
            <h2>Code intelligence - auto-indexing</h2>
            <p>
                Popular Go repositories will be indexed automatically via{' '}
                <a href="https://github.com/sourcegraph/lsif-go" target="_blank" rel="noreferrer noopener">
                    lsif-go
                </a>{' '}
                on{' '}
                <a href="https://sourcegraph.com" target="_blank" rel="noreferrer noopener">
                    Sourcegraph.com
                </a>
                . Enable precise code intelligence for non-Go code by{' '}
                <a
                    href="https://docs.sourcegraph.com/user/code_intelligence/lsif"
                    target="_blank"
                    rel="noreferrer noopener"
                >
                    uploading LSIF data
                </a>
                .
            </p>

            <FilteredConnection<GQL.ILSIFIndex, Omit<LsifIndexNodeProps, 'node'>>
                className="list-group list-group-flush mt-3"
                noun="index"
                pluralNoun="indexes"
                queryConnection={queryIndexes}
                nodeComponent={LsifIndexNode}
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
