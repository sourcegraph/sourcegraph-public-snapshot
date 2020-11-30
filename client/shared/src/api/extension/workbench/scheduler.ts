import { WorkbenchViewType } from 'sourcegraph'

export interface WorkbenchViewScheduler {
    schedule: (update: ViewUpdate) => void
    unsubscribe: () => void
}

let schedulerInstance: WorkbenchViewScheduler | undefined

// TODO(tj): Clarify language (update vs change)
interface ViewUpdate {
    type: 'creation' | 'update' | 'deletion'
    viewType: WorkbenchViewType
}

/**
 * Idempotent.
 *
 * @param onFlush Callback called when scheduler deems that it is appropriate to
 * notify the UI of view updates
 */
export function createWorkbenchViewScheduler(
    onFlush: (updatedViewTypes: Set<WorkbenchViewType>) => void,
    requestAnimationFrame: (callback: FrameRequestCallback) => number | undefined = globalThis.requestAnimationFrame
): WorkbenchViewScheduler {
    if (schedulerInstance) {
        return schedulerInstance
    }

    // TODO(tj): For now, we clone all views on each update, so we don't need anymore info from the update.
    // If we need to send deltas only (cloning views becomes too expensive), we will store whole updates
    // to pass to the `onFlush` handler
    let updatedViewTypes = new Set<WorkbenchViewType>()
    let updateScheduled = false
    let requestID: number | undefined

    function flush(): void {
        onFlush(updatedViewTypes)
        updatedViewTypes = new Set<WorkbenchViewType>()
        updateScheduled = false
    }

    schedulerInstance = {
        schedule: (update: ViewUpdate): void => {
            updatedViewTypes.add(update.viewType)

            if (!updateScheduled) {
                requestID = requestAnimationFrame(flush)
                updateScheduled = true
            }
        },
        unsubscribe: () => {
            if (requestID) {
                cancelAnimationFrame(requestID)
            }
            schedulerInstance = undefined
        },
    }

    return schedulerInstance
}
