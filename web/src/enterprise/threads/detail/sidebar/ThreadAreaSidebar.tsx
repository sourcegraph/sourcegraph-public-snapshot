import H from 'history'
import BellIcon from 'mdi-react/BellIcon'
import UserGroupIcon from 'mdi-react/UserGroupIcon'
import UserIcon from 'mdi-react/UserIcon'
import React from 'react'
import { Toggle } from '../../../../../../shared/src/components/Toggle'
import { ExtensionsControllerProps } from '../../../../../../shared/src/extensions/controller'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { CollapsibleSidebar } from '../../../../components/collapsibleSidebar/CollapsibleSidebar'
import { LabelIcon } from '../../../../projects/icons'
import { ThreadDeleteButton } from '../../form/ThreadDeleteButton'
import { ThreadSettings } from '../../settings'
import { CopyThreadLinkButton } from './CopyThreadLinkButton'

interface Props extends ExtensionsControllerProps {
    thread: GQL.IDiscussionThread
    onThreadUpdate: (thread: GQL.IDiscussionThread) => void
    threadSettings: ThreadSettings
    areaURL: string
    className?: string
    history: H.History
}

/**
 * The sidebar for the thread area (for a single thread).
 */
export const ThreadAreaSidebar: React.FunctionComponent<Props> = ({ thread, className = '', ...props }) => (
    <CollapsibleSidebar
        localStorageKey="thread-area__sidebar"
        side="right"
        className={`thread-area-sidebar d-flex flex-column border-left ${className}`}
        collapsedClassName="thread-area-sidebar--collapsed"
        expandedClassName="thread-area-sidebar--expanded"
    >
        {expanded => (
            <>
                <ul className="list-group list-group-flush px-2">
                    <li className="list-group-item thread-area-sidebar__item">
                        {expanded ? (
                            <>
                                <h6 className="font-weight-normal mb-0">Assignee</h6>
                                <div>
                                    <strong>@sqs</strong>
                                </div>
                            </>
                        ) : (
                            <UserIcon className="icon-inline" data-tooltip={`Assignee: @sqs`} />
                        )}
                    </li>
                    <li className="list-group-item thread-area-sidebar__item">
                        {expanded ? (
                            <>
                                <h6 className="font-weight-normal mb-0">Labels</h6>
                                <div>
                                    {thread.title
                                        .toLowerCase()
                                        .split(' ')
                                        .filter(w => w.length >= 3)
                                        .map((label, i) => (
                                            <span key={i} className={`badge mr-1 ${badgeColorClass(label)}`}>
                                                {label}
                                            </span>
                                        ))}
                                </div>
                            </>
                        ) : (
                            <LabelIcon className="icon-inline" data-tooltip="Labels TODO!(sqs)" />
                        )}
                    </li>
                    <li className="list-group-item thread-area-sidebar__item">
                        {expanded ? (
                            <>
                                <h6 className="font-weight-normal mb-0">3 participants</h6>
                                <div className="text-muted">@sqs @jtal3sf @rrono3 @jstykes</div>
                            </>
                        ) : (
                            <UserGroupIcon className="icon-inline" data-tooltip="TODO!(sqs)" />
                        )}
                    </li>
                    <li className="list-group-item thread-area-sidebar__item">
                        {expanded ? (
                            <h6 className="font-weight-normal mb-0 d-flex align-items-center justify-content-between">
                                Notifications <Toggle value={true} />
                            </h6>
                        ) : (
                            <BellIcon className="icon-inline" data-tooltip="TODO!(sqs)" />
                        )}
                    </li>
                    <li className="list-group-item thread-area-sidebar__item">
                        {expanded ? (
                            <h6 className="font-weight-normal mb-0 d-flex align-items-center justify-content-between">
                                Link
                                <CopyThreadLinkButton
                                    link={thread.url}
                                    className="btn btn-link btn-link-sm text-decoration-none px-0"
                                >
                                    #{thread.idWithoutKind}
                                </CopyThreadLinkButton>
                            </h6>
                        ) : (
                            <CopyThreadLinkButton
                                link={thread.url}
                                className="btn btn-link btn-link-sm text-decoration-none px-0"
                            />
                        )}
                    </li>
                    {expanded && (
                        <li className="list-group-item thread-area-sidebar__item">
                            <ThreadDeleteButton
                                {...props}
                                thread={thread}
                                buttonClassName="btn-link"
                                className="btn-sm px-0 text-decoration-none"
                                includeNounInLabel={true}
                            />
                        </li>
                    )}
                </ul>
            </>
        )}
    </CollapsibleSidebar>
)

function badgeColorClass(label: string): string {
    if (label === 'security' || label.endsWith('sec')) {
        return 'badge-danger'
    }
    const CLASSES = ['badge-primary', 'badge-warning', 'badge-info', 'badge-success', 'badge-danger']
    const k = label.split('').reduce((sum, c) => (sum += c.charCodeAt(0)), 0)
    return CLASSES[k % CLASSES.length]
}
