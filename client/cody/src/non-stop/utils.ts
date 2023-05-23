export enum CodyTaskState {
    'idle' = 0,
    'queued' = 1,
    'pending' = 2,
    'done' = 3,
    'applying' = 4,
    'error' = 5,
}

export type CodyTaskIcon = {
    [key in CodyTaskState]: {
        id: string
        icon: string
    }
}
/**
 * Icon for each task state
 */
export const fixupTaskIcon: CodyTaskIcon = {
    [CodyTaskState.idle]: {
        id: 'idle',
        icon: 'smiley',
    },
    [CodyTaskState.pending]: {
        id: 'pending',
        icon: 'sync~spin',
    },
    [CodyTaskState.done]: {
        id: 'done',
        icon: 'issue-closed',
    },
    [CodyTaskState.error]: {
        id: 'error',
        icon: 'stop',
    },
    [CodyTaskState.queued]: {
        id: 'queue',
        icon: 'clock',
    },
    [CodyTaskState.applying]: {
        id: 'applying',
        icon: 'sync~spin',
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
