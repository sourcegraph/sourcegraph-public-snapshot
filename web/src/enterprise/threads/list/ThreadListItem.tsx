import H from 'history'
import MessageOutlineIcon from 'mdi-react/MessageOutlineIcon'
import React from 'react'
import { Link } from 'react-router-dom'
import { LinkOrSpan } from '../../../../../shared/src/components/LinkOrSpan'
import { displayRepoName } from '../../../../../shared/src/components/RepoFileLink'
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { PlatformContextProps } from '../../../../../shared/src/platform/context'
import { ActorLink } from '../../../actor/ActorLink'
import { Timestamp } from '../../../components/time/Timestamp'
import { ThemeProps } from '../../../theme'
import { LabelableLabelsList } from '../../labels/labelable/LabelableLabelsList'
import { ThreadStateIcon } from '../common/threadState/ThreadStateIcon'
import { ThreadListContext } from './ThreadList'

export interface ThreadListItemContext {
    showRepository?: boolean

    /** A component rendered under the title. */
    itemSubtitle?: React.ComponentType<{
        thread: GQL.ThreadOrThreadPreview
    }>

    /** A component rendered on the right side.  */
    right?: React.ComponentType<
        {
            thread: GQL.ThreadOrThreadPreview
            location: H.Location
            history: H.History
        } & ExtensionsControllerProps &
            PlatformContextProps &
            ThemeProps
    >
}

interface Props
    extends ThreadListItemContext,
        ThreadListContext,
        ExtensionsControllerProps,
        PlatformContextProps,
        ThemeProps {
    thread: GQL.ThreadOrThreadPreview

    className?: string
    location: H.Location
    history: H.History
}

/**
 * An item in the list of threads.
 */
export const ThreadListItem: React.FunctionComponent<Props> = ({
    thread,
    showRepository,
    itemSubtitle: ItemSubtitle,
    right: Right,
    itemCheckboxes,
    className = '',
    ...props
}) => (
    <li className={`list-group-item ${className}`}>
        <div className="d-flex align-items-start">
            {itemCheckboxes && (
                <div
                    className="form-check ml-1 mr-2"
                    /* tslint:disable-next-line:jsx-ban-props */
                    style={{ marginTop: '4px' /* stylelint-disable-line declaration-property-unit-whitelist */ }}
                >
                    <input className="form-check-input position-static" type="checkbox" aria-label="Select item" />
                </div>
            )}
            <ThreadStateIcon
                thread={thread.__typename === 'Thread' ? thread : { kind: thread.kind, state: GQL.ThreadState.OPEN }}
                className="mt-1 mr-2"
            />
            <div className="flex-1">
                <div className="d-flex align-items-center flex-wrap">
                    <h4 className="d-flex align-items-center mb-0 mr-2">
                        <LinkOrSpan
                            to={thread.__typename === 'Thread' ? thread.url : undefined}
                            className="text-decoration-none text-body"
                        >
                            {thread.title}
                        </LinkOrSpan>
                        <span className="badge badge-secondary ml-1 d-none">123</span> {/* TODO!(sqs) */}
                    </h4>
                    {thread.__typename === 'Thread' && (
                        <LabelableLabelsList
                            labelable={thread}
                            showNoLabels={false}
                            showLoadingAndError={false}
                            className="d-flex align-items-center ml-2"
                            itemClassName="mr-2 py-1"
                        />
                    )}
                </div>
                <ul className="list-unstyled d-flex align-items-center small text-muted mb-0">
                    <li>
                        {thread.__typename === 'Thread' ? (
                            <span className="text-muted mr-2">
                                {showRepository && (
                                    <Link to={thread.repository.url} className="text-muted">
                                        {displayRepoName(thread.repository.name)}
                                    </Link>
                                )}
                                #{thread.number}
                            </span>
                        ) : (
                            <span className="text-muted mr-2">
                                Create new {thread.kind.toLowerCase()}{' '}
                                {showRepository && (
                                    <>
                                        in{' '}
                                        <Link to={thread.repository.url}>
                                            {displayRepoName(thread.repository.name)}
                                        </Link>
                                    </>
                                )}
                            </span>
                        )}
                    </li>
                    {thread.__typename === 'Thread' && (
                        <li>
                            created <Timestamp date={thread.createdAt} /> by <ActorLink actor={thread.author} />
                        </li>
                    )}
                    {thread.assignees.nodes.length > 0 && (
                        <li>
                            &bull; Assigned to{' '}
                            {thread.assignees.nodes.map((assignee, i) => (
                                <ActorLink key={i} actor={assignee} className="mr-1" />
                            ))}
                        </li>
                    )}
                    {ItemSubtitle && (
                        <li className="ml-2">
                            &bull; <ItemSubtitle thread={thread} />
                        </li>
                    )}
                    {/* TODO!(sqs): show contents {thread.targets.totalCount > 0 && (
                        <li className="ml-2 d-flex align-items-center">
                            <FileFindIcon className="icon-inline mr-1" /> {thread.targets.totalCount}{' '}
                            {pluralize('item', thread.targets.totalCount)}
                        </li>
                    )}*/}
                </ul>
            </div>
            <div>
                <ul className="list-inline d-flex align-items-center">
                    {Right && <Right {...props} thread={thread} />}
                    {thread.__typename === 'Thread' && thread.comments.totalCount >= 1 && (
                        <li className="list-inline-item">
                            <small className="text-muted">
                                <MessageOutlineIcon className="icon-inline" /> {thread.comments.totalCount - 1}
                            </small>
                        </li>
                    )}
                </ul>
            </div>
        </div>
    </li>
)
