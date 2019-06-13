import FileDocumentBoxMultipleIcon from 'mdi-react/FileDocumentBoxMultipleIcon'
import InboxIcon from 'mdi-react/InboxIcon'
import PlayCircleIcon from 'mdi-react/PlayCircleIcon'
import SettingsIcon from 'mdi-react/SettingsIcon'
import React from 'react'
import { NavLink } from 'react-router-dom'
import { ChatIcon } from '../../../../../shared/src/components/icons'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { ThreadSettings } from '../settings'

interface Props {
    thread: GQL.IDiscussionThread
    threadSettings: ThreadSettings
    sections: { review: boolean; changes: boolean; actions: boolean; settings: boolean }
    areaURL: string
    className?: string
}

const isHandled = (item: GQL.IDiscussionThreadTargetRepo, threadSettings: ThreadSettings): boolean =>
    (threadSettings.pullRequests || []).some(pull => pull.items.includes(item.id))

const NAV_LINK_CLASS_NAME = 'thread-area-navbar__nav-link nav-link rounded-0 px-4'

/**
 * The navbar for a single thread.
 */
export const ThreadAreaNavbar: React.FunctionComponent<Props> = ({
    thread,
    threadSettings,
    sections,
    areaURL,
    className = '',
}) => (
    <div className={`thread-area-navbar border-bottom ${className}`}>
        <div className="container px-0">
            <div className="nav nav-pills flex-nowrap">
                <div className="nav-item">
                    <NavLink
                        to={areaURL}
                        exact={true}
                        className={NAV_LINK_CLASS_NAME}
                        activeClassName="thread-area-navbar__nav-link--active"
                    >
                        <ChatIcon className="icon-inline" /> Conversation{' '}
                        <span className="badge badge-secondary">{thread.comments.totalCount - 1}</span>
                    </NavLink>
                </div>
                {sections.review && (
                    <div className="nav-item">
                        <NavLink
                            to={`${areaURL}/inbox`}
                            className={NAV_LINK_CLASS_NAME}
                            activeClassName="thread-area-navbar__nav-link--active"
                        >
                            <InboxIcon className="icon-inline" /> Items{' '}
                            <span className="badge badge-secondary d-none">
                                {
                                    thread.targets.nodes
                                        .filter(
                                            (v): v is GQL.IDiscussionThreadTargetRepo =>
                                                v.__typename === 'DiscussionThreadTargetRepo'
                                        )
                                        .filter(v => !v.isIgnored)
                                        .filter(v => !isHandled(v, threadSettings)).length
                                }
                            </span>
                        </NavLink>
                    </div>
                )}
                {sections.changes && (
                    <div className="nav-item">
                        <NavLink
                            to={`${areaURL}/changes`}
                            className={NAV_LINK_CLASS_NAME}
                            activeClassName="thread-area-navbar__nav-link--active"
                        >
                            <FileDocumentBoxMultipleIcon className="icon-inline" /> Changes
                        </NavLink>
                    </div>
                )}
                {sections.actions && (
                    <div className="nav-item">
                        <NavLink
                            to={`${areaURL}/actions`}
                            className={NAV_LINK_CLASS_NAME}
                            activeClassName="thread-area-navbar__nav-link--active"
                        >
                            <PlayCircleIcon className="icon-inline" /> Actions
                        </NavLink>
                    </div>
                )}
                <div className="flex-1" />
                {sections.settings && (
                    <div className="nav-item">
                        <NavLink
                            to={`${areaURL}/settings`}
                            className={NAV_LINK_CLASS_NAME}
                            activeClassName="thread-area-navbar__nav-link--active"
                        >
                            <SettingsIcon className="icon-inline" />
                        </NavLink>
                    </div>
                )}
            </div>
        </div>
    </div>
)
