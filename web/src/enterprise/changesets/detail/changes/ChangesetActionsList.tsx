import React from 'react'
import { ExtensionsControllerProps } from '../../../../../../shared/src/extensions/controller'
import * as GQL from '../../../../../../shared/src/graphql/schema'
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
 * A list of actions run in a changeset.
 */
export const ChangesetActionsList: React.FunctionComponent<Props> = ({ threadSettings, className = '' }) =>
    threadSettings.changesetActionDescriptions && threadSettings.changesetActionDescriptions.length > 0 ? (
        <div className={`changeset-actions-list ${className}`}>
            <ul className="list-group">
                {threadSettings.changesetActionDescriptions.map((a, i) => (
                    <li key={i} className="list-group-item d-flex align-items-start">
                        <ActionsIcon className="icon-inline small mr-2" />
                        <header>
                            <h6 className="mb-0 font-size-base mr-4">{a.title}</h6>
                            {a.detail && <span className="text-muted">{a.detail}</span>}
                        </header>
                        <div className="flex-1"></div>
                        <small className="text-muted">
                            <strong>@{a.user}</strong> <Timestamp date={a.timestamp} noAbout={true} />
                        </small>
                    </li>
                ))}
            </ul>
        </div>
    ) : null
