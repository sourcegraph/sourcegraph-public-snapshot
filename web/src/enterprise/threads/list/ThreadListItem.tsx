import H from 'history'
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
import { AbstractThreadListItem } from './AbstractThreadListItem'
import { isDefined } from '../../../../../shared/src/util/types'

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
    <AbstractThreadListItem
        left={<ThreadStateIcon thread={thread} />}
        title={
            <LinkOrSpan
                to={
                    thread.__typename === 'Thread' && thread.externalURLs && thread.externalURLs.length > 0
                        ? thread.externalURLs[0].url
                        : undefined
                }
                target="_blank"// TODO!(sqs): remove target=_blank when threads have their own pages on sourcegraph
                rel="noopener noreferrer"
                className="text-decoration-none text-body"
            >
                {thread.title}
            </LinkOrSpan>
        }
        afterTitle={
            thread.__typename === 'Thread' && (
                <LabelableLabelsList
                    labelable={thread}
                    showNoLabels={false}
                    className="d-flex align-items-center ml-2"
                    itemClassName="mr-2 py-1"
                />
            )
        }
        detail={[
            thread.__typename === 'Thread' && (
                <span className="text-muted mr-2">
                    {showRepository && (
                        <Link to={thread.repository.url} className="text-muted">
                            {displayRepoName(thread.repository.name)}
                        </Link>
                    )}
                    #TODO
                </span>
            ),
            thread.__typename === 'Thread' && (
                <>
                    created <Timestamp date={thread.createdAt} /> by TODO(add author)
                </>
            ),
            thread.assignees.nodes.length > 0 && (
                <>
                    &bull; Assigned to{' '}
                    {thread.assignees.nodes.map((assignee, i) => (
                        // eslint-disable-next-line react/no-array-index-key
                        <ActorLink key={i} actor={assignee} className="mr-1" />
                    ))}
                </>
            ),
            ItemSubtitle && (
                <span className="ml-2">
                    &bull; <ItemSubtitle thread={thread} />
                </span>
            ),
        ].filter(isDefined)}
        right={[Right && <Right key={0} {...props} thread={thread} />].filter(isDefined)}
        className={className}
    />
)
