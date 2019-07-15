import SyncIcon from 'mdi-react/SyncIcon'
import React from 'react'
import * as sourcegraph from 'sourcegraph'
import { displayRepoName } from '../../../../../../shared/src/components/RepoFileLink'
import { ExtensionsControllerProps } from '../../../../../../shared/src/extensions/controller'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { isDefined } from '../../../../../../shared/src/util/types'
import { parseRepoURI } from '../../../../../../shared/src/util/url'
import { Timestamp } from '../../../../components/time/Timestamp'
import { ActionsIcon } from '../../../../util/octicons'
import { ThreadSettings } from '../../../threads/settings'

interface Props extends ExtensionsControllerProps {
    thread: GQL.IDiscussionThread
    xchangeset: GQL.IChangeset
    threadSettings: ThreadSettings

    className?: string
}

/**
 * A list of operations performed by a changeset.
 */
export const ChangesetOperationsList: React.FunctionComponent<Props> = ({ thread, threadSettings, className = '' }) => (
    <div className={`changeset-operations-list ${className}`}>
        {threadSettings.plan && threadSettings.plan.operations.length > 0 ? (
            <>
                <div className="border border-success p-3 mb-4 d-flex align-items-stretch">
                    <SyncIcon className="flex-0 mr-2" />{' '}
                    <div className="flex-1">
                        <h4 className="mb-0">Automatic updates enabled</h4>
                        <p className="mb-0">
                            The operations {thread.status === GQL.ThreadStatus.PREVIEW ? 'will' : ''} run when any base
                            branch changes or when a new repository matches.
                        </p>
                    </div>
                </div>
                <ul className="list-group">
                    {threadSettings.plan.operations.map((op, i) => (
                        <li key={i} className="list-group-item d-flex align-items-start">
                            <ActionsIcon className="icon-inline small mr-2" />
                            <header>
                                <h6 className="mb-0 font-size-base font-weight-normal mr-4">{op.message}</h6>
                            </header>
                            <div className="flex-1"></div>
                            {op.diagnostics && (
                                <small className="text-muted mt-1">{diagnosticQueryLabel(op.diagnostics)}</small>
                            )}
                        </li>
                    ))}
                </ul>
            </>
        ) : (
            <span className="text-muted">No operations</span>
        )}
    </div>
)

function diagnosticQueryLabel(query: sourcegraph.DiagnosticQuery): string {
    const parts = [
        query.type && `${query.type} diagnostics`,
        query.tag && `tagged with '${query.tag}'`,
        query.document &&
            query.document.pattern &&
            `in repository ${displayRepoName(parseRepoURI(query.document.pattern).repoName)} ${
                parseRepoURI(query.document.pattern).filePath
                    ? `file ${parseRepoURI(query.document.pattern).filePath}`
                    : ''
            }`,
    ]
        .filter(isDefined)
        .join(' ')
    if (parts === '') {
        return 'All diagnostics'
    }
    return `Fixes ${parts}`
}
