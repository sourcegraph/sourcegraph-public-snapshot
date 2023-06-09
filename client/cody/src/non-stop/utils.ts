export enum CodyTaskState {
    'idle' = 0,
    'queued' = 1,
    'asking' = 2,
    'marking' = 3,
    'ready' = 4,
    'applying' = 5,
    'fixed' = 6,
    'error' = 7,
}

export type CodyTaskList = {
    [key in CodyTaskState]: {
        id: string
        icon: string
        description: string
    }
}
/**
 * Icon for each task state
 */
export const fixupTaskList: CodyTaskList = {
    [CodyTaskState.idle]: {
        id: 'idle',
        icon: 'clock',
        description: 'Initial state - all task starts from here',
    },
    [CodyTaskState.asking]: {
        id: 'asking',
        icon: 'sync~spin',
        description: 'In the process of pending for Cody response',
    },
    [CodyTaskState.ready]: {
        id: 'ready',
        icon: 'smiley',
        description: 'Cody has responsed with suggestions and is ready to apply them',
    },
    [CodyTaskState.error]: {
        id: 'error',
        icon: 'stop',
        description: 'The task has been completed and returned error',
    },
    [CodyTaskState.queued]: {
        id: 'queue',
        icon: 'debug-pause',
        description: 'The task is in the queue to be processed by Cody',
    },
    [CodyTaskState.applying]: {
        id: 'applying',
        icon: 'sync~spin',
        description: 'In the process of applying the fixups to the docs',
    },
    [CodyTaskState.marking]: {
        id: 'marking',
        icon: 'edit',
        description: 'Marking the fixups as Cody responses',
    },
    [CodyTaskState.fixed]: {
        id: 'fixed',
        icon: 'pass-filled',
        description: 'Suggestions from Cody have been applied to the docs successfully',
    },
}
/**
 * Get the last part of the file path after the last slash
 */
export function getFileNameAfterLastDash(filePath: string): string {
    const lastDashIndex = filePath.lastIndexOf('/')
    if (lastDashIndex === -1) {
        return filePath
    }
    return filePath.slice(lastDashIndex + 1)
}
