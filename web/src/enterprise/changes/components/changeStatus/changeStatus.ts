import * as GQL from '../../../../../../shared/src/graphql/schema'
import { ChangesIcon } from '../../icons'

/**
 * The subset of {@link GQL.IChange}'s that is needed for displaying the change's status.
 */
export interface ChangeWithStatus extends Pick<GQL.IChange, 'status'> {}

type ChangeStatusColor = 'success' | 'info' | 'danger'

const STATUS_COLOR: Record<GQL.ChangeStatus, ChangeStatusColor> = {
    OPEN: 'success',
    CLOSED: 'danger',
}

const STATUS_TEXT: Record<GQL.ChangeStatus, string> = {
    OPEN: 'Open',
    CLOSED: 'Closed',
}

/**
 * Returns information about how to display the change's status.
 */
export const changeStatusInfo = (
    change: ChangeWithStatus
): {
    color: ChangeStatusColor
    icon: React.ComponentType<{ className?: string }>
    text: string
} => ({
    color: STATUS_COLOR[change.status],
    icon: ChangesIcon,
    text: STATUS_TEXT[change.status],
})
