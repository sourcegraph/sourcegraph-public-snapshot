import * as GQL from '../../../../../../shared/src/graphql/schema'
import { GitPullRequestIcon } from '../../../../util/octicons'
import { ChecksIcon } from '../../../checks/icons'
import { ThreadsIcon } from '../../../threads/icons'

/**
 * The subset of {@link GQL.IDiscussionThread}'s that is needed for displaying the thread's status icons.
 */
export interface ThreadStatusFields extends Pick<GQL.IDiscussionThread, 'status' | 'type'> {}

type ThreadStatusColor = 'success' | 'danger' | 'info' | 'secondary'

const STATUS_COLOR: Record<GQL.ThreadStatus, ThreadStatusColor> = {
    OPEN: 'success',
    CLOSED: 'danger',
    PREVIEW: 'secondary',
}

const threadIcon = (thread: ThreadStatusFields): React.ComponentType<{ className?: string }> =>
    thread.type === GQL.ThreadType.ISSUE
        ? ChecksIcon
        : thread.type === GQL.ThreadType.CHANGESET
        ? GitPullRequestIcon
        : ThreadsIcon

const statusText = (thread: ThreadStatusFields) => {
    switch (thread.status) {
        case GQL.ThreadStatus.OPEN:
            return 'Open'
        case GQL.ThreadStatus.CLOSED:
            return 'Closed'
        case GQL.ThreadStatus.PREVIEW:
            return 'Preview'
    }
}

const threadTooltip = (thread: ThreadStatusFields): string =>
    `${statusText(thread)} ${thread.type === GQL.ThreadType.ISSUE ? 'check' : 'thread'}`

/**
 * Returns information about how to display the thread's status.
 */
export const threadStatusInfo = (
    thread: ThreadStatusFields
): {
    color: ThreadStatusColor
    icon: React.ComponentType<{ className?: string }>
    text: string
    tooltip: string
} => ({
    color: STATUS_COLOR[thread.status],
    icon: threadIcon(thread),
    text: statusText(thread),
    tooltip: threadTooltip(thread),
})
