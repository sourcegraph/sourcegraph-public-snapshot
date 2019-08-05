import * as GQL from '../../../../../shared/src/graphql/schema'
import { ChangesetsIcon } from '../../changesets/icons'
import { ThreadsIcon } from '../../threads/icons'

/**
 * The subset of fields that is needed for displaying the thread's state icons.
 */
export interface ThreadStateFields extends Pick<GQL.ThreadOrIssueOrChangeset, '__typename' | 'state'> {}

type ThreadStateColor = 'success' | 'danger' | 'info' | 'secondary' | 'purple'

const COLOR: Record<GQL.ThreadOrIssueOrChangeset['state'], ThreadStateColor> = {
    OPEN: 'success',
    CLOSED: 'danger',
    MERGED: 'purple', // TODO!(sqs): make purple
}

const icon = (thread: ThreadStateFields): React.ComponentType<{ className?: string }> =>
    thread.__typename === 'Changeset' ? ChangesetsIcon : ThreadsIcon

const text = (thread: ThreadStateFields) => {
    switch (thread.state) {
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

const tooltip = (thread: ThreadStateFields): string => `${text(thread)} ${thread.__typename.toLowerCase()}`

/**
 * Returns information about how to display the thread's state.
 */
export const threadStateInfo = (
    thread: ThreadStateFields
): {
    color: ThreadStateColor
    icon: React.ComponentType<{ className?: string }>
    text: string
    tooltip: string
} => ({
    color: COLOR[thread.state],
    icon: icon(thread),
    text: text(thread),
    tooltip: tooltip(thread),
})
