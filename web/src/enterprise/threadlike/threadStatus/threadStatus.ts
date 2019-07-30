import * as GQL from '../../../../../shared/src/graphql/schema'
import { ChangesetsIcon } from '../../changesets/icons'
import { ThreadsIcon } from '../../threads/icons'

/**
 * The subset of fields that is needed for displaying the thread's status icons.
 */
export interface ThreadStatusFields extends Pick<GQL.ThreadOrIssueOrChangeset, '__typename' | 'status'> {}

type ThreadStatusColor = 'success' | 'danger' | 'info' | 'secondary'

const STATUS_COLOR: Record<GQL.ThreadOrIssueOrChangeset['status'], ThreadStatusColor> = {
    OPEN: 'success',
    CLOSED: 'danger',
    MERGED: 'info', // TODO!(sqs): make purple
}

const threadIcon = (thread: ThreadStatusFields): React.ComponentType<{ className?: string }> =>
    thread.__typename === 'Changeset' ? ChangesetsIcon : ThreadsIcon

const statusText = (thread: ThreadStatusFields) => {
    switch (thread.status) {
        case GQL.ThreadStatus.OPEN:
        case GQL.ChangesetStatus.OPEN:
            return 'Open'
        case GQL.ThreadStatus.CLOSED:
        case GQL.ChangesetStatus.CLOSED:
            return 'Closed'
        case GQL.ChangesetStatus.MERGED:
            return 'Merged'
    }
}

const threadTooltip = (thread: ThreadStatusFields): string => `${statusText(thread)} ${thread.__typename.toLowerCase()}`

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
