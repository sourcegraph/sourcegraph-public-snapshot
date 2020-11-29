import { WorkbenchViewType } from 'sourcegraph'

export interface WorkbenchViewScheduler {
    schedule: (update: ViewUpdate) => void
}

let schedulerInstance: WorkbenchViewScheduler | undefined

// TODO(tj): Clarify language (update vs change)
interface ViewUpdate {
    type: 'creation' | 'update' | 'deletion'
    viewType: WorkbenchViewType
}

export function createWorkbenchViewScheduler(
    // TODO(tj): document
    onFlush: (updatedViewTypes: Set<WorkbenchViewType>) => void
): WorkbenchViewScheduler {
    if (schedulerInstance) {
        return schedulerInstance
    }

    // TODO: if we don't need anymore info from the update, just make it a Set of updated types
    let updatedViewTypes = new Set<WorkbenchViewType>()
    let updateScheduled = false

    function flush(): void {
        onFlush(updatedViewTypes)
        updatedViewTypes = new Set<WorkbenchViewType>()
        updateScheduled = false
    }

    schedulerInstance = {
        schedule: (update: ViewUpdate): void => {
            updatedViewTypes.add(update.viewType)

            if (!updateScheduled) {
                requestAnimationFrame(flush)
                updateScheduled = true
            }
        },
    }

    return schedulerInstance
}
