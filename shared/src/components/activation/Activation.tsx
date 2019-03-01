import H from 'history'
import { LinkProps } from '../Link'

/**
 * Represents the activation status of the current user.
 */
export interface Activation<K extends string> {
    /**
     * The steps that make up the activation list
     */
    steps: ActivationStep<K>[]

    /**
     * The completion status of each activation step
     */
    completed?: { [key in K]: boolean }

    /**
     * Updates the activation status with the given steps and their completion status.
     */
    update: (u: { [key in K]: boolean }) => void

    /**
     * Resync the activation status from the server.
     */
    refetch: () => void
}

/**
 * Component props should inherit from this to include activation status.
 */
export interface ActivationProps<K extends string> {
    activation?: Activation<K>
}

/**
 * One step in the activation status.
 */
export interface ActivationStep<K> {
    id: K
    title: string
    detail: string
    link?: LinkProps
    onClick?: (event: React.MouseEvent<HTMLElement>, history: H.History) => void
}

/**
 * Returns the percent of activation checklist items completed.
 */
export function percentageDone<K extends string>(info?: { [K: string]: boolean }): number {
    if (!info) {
        return 0
    }
    const vals = Object.values(info)
    return (100 * vals.filter(e => e).length) / vals.length
}
