export enum CodyTaskState {
    'idle' = 0,
    'pending' = 1,
    'done' = 2,
    'error' = 3,
    'stopped' = 4,
    'queued' = 5,
}

export type CodyTaskIcon = {
    [key in CodyTaskState]: {
        id: string
        icon: string
    }
}

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
    [CodyTaskState.stopped]: {
        id: 'removed',
        icon: 'circle-slash',
    },
}
