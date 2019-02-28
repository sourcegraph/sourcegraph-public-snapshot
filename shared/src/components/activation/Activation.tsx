import H from 'history'

/**
 * Represents the activation status of the current user.
 */
export interface Activation {
    /**
     * The steps that make up the activation list
     */
    steps: ActivationStep[]

    /**
     * The completion status of each activation step
     */
    completed?: { [key: string]: boolean }

    /**
     * Updates the activation status with the given steps and their completion status.
     */
    update: (u: { [key: string]: boolean }) => void

    /**
     * Resync the activation status from the server.
     */
    refetch: () => void
}

/**
 * Component props should inherit from this to include activation status.
 */
export interface ActivationProps {
    activation?: Activation
}

/**
 * One step in the activation status.
 */
export interface ActivationStep {
    id: string
    title: string
    detail: string
    action: (h: H.History) => void
}

/**
 * Returns the percent of activation checklist items completed.
 */
export const percentageDone = (info?: { [key: string]: boolean }): number => {
    if (!info) {
        return 0
    }
    const vals = Object.values(info)
    return (100 * vals.filter(e => e).length) / vals.length
}
