export enum CodyTaskState {
    'idle' = 0,
    'waiting' = 1,
    'asking' = 2,
    'ready' = 3,
    'applying' = 4,
    'fixed' = 5,
    'error' = 6,
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
        description: 'Initial state',
    },
    [CodyTaskState.waiting]: {
        id: 'waiting',
        // TODO: Per Figma, this should be a Cody wink
        icon: 'debug-pause',
        description: 'The task is waiting to be processed by Cody',
    },
    [CodyTaskState.asking]: {
        id: 'asking',
        icon: 'sync~spin',
        description: 'Cody is preparing a response',
    },
    [CodyTaskState.ready]: {
        id: 'ready',
        icon: 'pencil',
        description: 'Cody has responsed with suggestions and is ready to apply them',
    },
    [CodyTaskState.applying]: {
        id: 'applying',
        icon: 'pencil',
        description: 'The fixup is being applied to the document',
    },
    [CodyTaskState.fixed]: {
        id: 'fixed',
        icon: 'pass-filled',
        description: 'Suggestions from Cody have been applied or discarded',
    },
    [CodyTaskState.error]: {
        id: 'error',
        icon: 'stop',
        description: 'The task failed',
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
