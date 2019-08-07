import * as GQL from '../../../../../../shared/src/graphql/schema'
import { ChangesetsIcon, IssuesIcon, ThreadsIcon } from '../../icons'

/**
 * The subset of fields that is needed for displaying the thread's state icons.
 */
export interface ThreadStateFields extends Pick<GQL.IThread, 'state' | 'kind'> {}

type ThreadStateColor = 'success' | 'danger' | 'info' | 'secondary' | 'purple'

const COLOR: Record<GQL.IThread['state'], ThreadStateColor> = {
    OPEN: 'success',
    CLOSED: 'danger',
    MERGED: 'purple', // TODO!(sqs): make purple
}

const ICON: Record<GQL.ThreadKind, React.ComponentType<{ className?: string }>> = {
    [GQL.ThreadKind.DISCUSSION]: ThreadsIcon,
    [GQL.ThreadKind.ISSUE]: IssuesIcon,
    [GQL.ThreadKind.CHANGESET]: ChangesetsIcon,
}

const text = (thread: ThreadStateFields) => {
    switch (thread.state) {
        case GQL.ThreadState.OPEN:
            return 'Open'
        case GQL.ThreadState.MERGED:
            return 'Merged'
        case GQL.ThreadState.CLOSED:
            return 'Closed'
    }
}

const tooltip = (thread: ThreadStateFields): string => `${text(thread)} ${thread.kind.toLowerCase()}`

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
    icon: ICON[thread.kind],
    text: text(thread),
    tooltip: tooltip(thread),
})
