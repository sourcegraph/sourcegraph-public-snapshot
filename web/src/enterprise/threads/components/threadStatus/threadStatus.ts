import * as GQL from '../../../../../../shared/src/graphql/schema'
import { ChecksIcon } from '../../../checks/icons'
import { ThreadsIcon } from '../../icons'

/**
 * The subset of {@link GQL.IDiscussionThread}'s that is needed for displaying the thread's status icons.
 */
export interface ThreadStatusFields extends Pick<GQL.IDiscussionThread, 'status' | 'type'> {}

type ThreadStatusColor = 'success' | 'danger' | 'info'

const STATUS_COLOR: Record<GQL.ThreadStatus, ThreadStatusColor> = {
    OPEN_ACTIVE: 'success',
    INACTIVE: 'info',
    CLOSED: 'danger',
}

const threadIcon = (thread: ThreadStatusFields): React.ComponentType<{ className?: string }> =>
    thread.type === GQL.ThreadType.CHECK ? ChecksIcon : ThreadsIcon

const statusText = (thread: ThreadStatusFields) => {
    switch (thread.status) {
        case GQL.ThreadStatus.OPEN_ACTIVE:
            return thread.type === GQL.ThreadType.CHECK ? 'Active' : 'Open'
        case GQL.ThreadStatus.INACTIVE:
            return 'Inactive'
        case GQL.ThreadStatus.CLOSED:
            return 'Closed'
    }
}

const threadTooltip = (thread: ThreadStatusFields): string =>
    `${statusText(thread)} ${thread.type === GQL.ThreadType.CHECK ? 'check' : 'thread'}`

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
