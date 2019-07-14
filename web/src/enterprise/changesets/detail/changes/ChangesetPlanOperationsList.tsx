import React from 'react'
import { ExtensionsControllerProps } from '../../../../../../shared/src/extensions/controller'
import * as GQL from '../../../../../../shared/src/graphql/schema'
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
export const ChangesetPlanOperationsList: React.FunctionComponent<Props> = ({ threadSettings, className = '' }) =>
    threadSettings.plan && threadSettings.plan.operations.length > 0 ? (
        <div className={`changeset-actions-list ${className}`}>
            <ul className="list-group">
                {threadSettings.plan.operations.map((op, i) => (
                    <li key={i} className="list-group-item d-flex align-items-start">
                        <ActionsIcon className="icon-inline small mr-2" />
                        <header>
                            <h6 className="mb-0 font-size-base font-weight-normal mr-4">{op.message}</h6>
                        </header>
                        <div className="flex-1"></div>
                        {op.diagnostics && (
                            <small className="text-muted mt-1">
                                {op.diagnostics.document &&
                                    op.diagnostics.document.pattern &&
                                    parseRepoURI(op.diagnostics.document.pattern).filePath}
                            </small>
                        )}
                    </li>
                ))}
            </ul>
        </div>
    ) : null
