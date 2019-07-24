import BellIcon from 'mdi-react/BellIcon'
import UserGroupIcon from 'mdi-react/UserGroupIcon'
import UserIcon from 'mdi-react/UserIcon'
import React from 'react'
import { Toggle } from '../../../../../shared/src/components/Toggle'
import { CollapsibleSidebar } from '../../../components/collapsibleSidebar/CollapsibleSidebar'
import { LabelIcon } from '../../../projects/icons'
import { ThreadlikeAreaContext } from '../ThreadlikeArea'

interface Props extends Pick<ThreadlikeAreaContext, 'thread'> {
    className?: string
}

/**
 * The sidebar for a single threadlike.
 */
export const ThreadlikeAreaSidebar: React.FunctionComponent<Props> = ({ thread, className = '' }) => (
    <CollapsibleSidebar
        localStorageKey="threadlike-area__sidebar"
        side="right"
        className={`threadlike-area-sidebar d-flex flex-column border-left ${className}`}
        collapsedClassName="threadlike-area-sidebar--collapsed"
        expandedClassName="threadlike-area-sidebar--expanded"
    >
        {expanded => (
            <>
                <ul className="list-group list-group-flush px-2">
                    <li className="list-group-item threadlike-area-sidebar__item">
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
                    <li className="list-group-item threadlike-area-sidebar__item">
                        {expanded ? (
                            <>
                                <h6 className="font-weight-normal mb-0">Labels</h6>
                                <div>
                                    {thread.title
                                        .toLowerCase()
                                        .split(' ')
                                        .filter(w => w.length >= 3)
                                        .map((label, i) => (
                                            <span key={i} className={`badge mr-1 TODO!(sqs)`}>
                                                {label}
                                            </span>
                                        ))}
                                </div>
                            </>
                        ) : (
                            <LabelIcon className="icon-inline" data-tooltip="Labels TODO!(sqs)" />
                        )}
                    </li>
                    <li className="list-group-item threadlike-area-sidebar__item">
                        {expanded ? (
                            <>
                                <h6 className="font-weight-normal mb-0">3 participants</h6>
                                <div className="text-muted">@sqs @jtal3sf @rrono3 @jstykes</div>
                            </>
                        ) : (
                            <UserGroupIcon className="icon-inline" data-tooltip="TODO!(sqs)" />
                        )}
                    </li>
                    <li className="list-group-item threadlike-area-sidebar__item">
                        {expanded ? (
                            <h6 className="font-weight-normal mb-0 d-flex align-items-center justify-content-between">
                                Notifications <Toggle value={true} />
                            </h6>
                        ) : (
                            <BellIcon className="icon-inline" data-tooltip="TODO!(sqs)" />
                        )}
                    </li>
                    {/* TODO!(sqs) <li className="list-group-item threadlike-area-sidebar__item">
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
                        </li>*/}
                    {/* TODO!(sqs) expanded && (
                        <li className="list-group-item threadlike-area-sidebar__item">
                            <ThreadDeleteButton
                                {...props}
                                thread={thread}
                                buttonClassName="btn-link"
                                className="btn-sm px-0 text-decoration-none"
                                includeNounInLabel={true}
                            />
                        </li>
                    )*/}
                </ul>
            </>
        )}
    </CollapsibleSidebar>
)
