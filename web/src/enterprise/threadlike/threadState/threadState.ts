import * as GQL from '../../../../../shared/src/graphql/schema'
import { ChangesetsIcon } from '../../changesets/icons'
import { ThreadsIcon } from '../../threads/icons'

/**
 * The subset of fields that is needed for displaying the thread's status icons.
 */
export interface ThreadStateFields extends Pick<GQL.ThreadOrIssueOrChangeset, '__typename' | 'status'> {}

type ThreadStateColor = 'success' | 'danger' | 'info' | 'secondary'

const STATUS_COLOR: Record<GQL.ThreadOrIssueOrChangeset['status'], ThreadStateColor> = {
    OPEN: 'success',
    CLOSED: 'danger',
    MERGED: 'info', // TODO!(sqs): make purple
}

const threadIcon = (thread: ThreadStateFields): React.ComponentType<{ className?: string }> =>
    thread.__typename === 'Changeset' ? ChangesetsIcon : ThreadsIcon

const statusText = (thread: ThreadStateFields) => {
    switch (thread.status) {
        case GQL.ThreadState.OPEN:
        case GQL.ChangesetState.OPEN:
            return 'Open'
        case GQL.ThreadState.CLOSED:
        case GQL.ChangesetState.CLOSED:
            return 'Closed'
        case GQL.ChangesetState.MERGED:
            return 'Merged'
    }
}

const threadTooltip = (thread: ThreadStateFields): string => `${statusText(thread)} ${thread.__typename.toLowerCase()}`

/**
 * Returns information about how to display the thread's status.
 */
export const threadStateInfo = (
    thread: ThreadStateFields
): {
    color: ThreadStateColor
    icon: React.ComponentType<{ className?: string }>
    text: string
    tooltip: string
} => ({
    color: STATUS_COLOR[thread.status],
    icon: threadIcon(thread),
    text: statusText(thread),
    tooltip: threadTooltip(thread),
})
