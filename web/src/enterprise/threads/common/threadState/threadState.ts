import AlertCircleOutlineIcon from 'mdi-react/AlertCircleOutlineIcon'
import CommentTextMultipleIcon from 'mdi-react/CommentTextMultipleIcon'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { GitPullRequestIcon } from '../../../../util/octicons'

/**
 * The subset of fields that is needed for displaying the thread's state icons.
 */
export type ThreadStateFields =
    | Pick<GQL.IThread, '__typename' | 'state' | 'kind' | 'isDraft'>
    | Pick<GQL.IThreadPreview, '__typename' | 'kind' | 'isDraft'>

type ThreadStateColor = 'success' | 'danger' | 'info' | 'secondary' | 'purple'

const COLOR: Record<GQL.IThread['state'], ThreadStateColor> = {
    OPEN: 'success',
    CLOSED: 'danger',
    MERGED: 'purple', // TODO!(sqs): make purple
}

const ICON: Record<GQL.ThreadKind, React.ComponentType<{ className?: string }>> = {
    [GQL.ThreadKind.DISCUSSION]: CommentTextMultipleIcon,
    [GQL.ThreadKind.ISSUE]: AlertCircleOutlineIcon,
    [GQL.ThreadKind.CHANGESET]: GitPullRequestIcon,
}

const text = (thread: ThreadStateFields): string => {
    switch (thread.__typename) {
        case 'Thread':
            switch (thread.state) {
                case GQL.ThreadState.OPEN:
                    return 'Open'
                case GQL.ThreadState.MERGED:
                    return 'Merged'
                case GQL.ThreadState.CLOSED:
                    return 'Closed'
            }
        case 'ThreadPreview':
            return 'Open'
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
    color: thread.isDraft ? 'secondary' : COLOR[thread.__typename === 'Thread' ? thread.state : GQL.ThreadState.OPEN],
    icon: ICON[thread.kind],
    text: `${text(thread)}${thread.isDraft ? ' (draft)' : ''}`,
    tooltip: tooltip(thread),
})
