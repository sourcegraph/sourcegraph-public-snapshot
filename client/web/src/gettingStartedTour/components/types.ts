export enum TourLanguage {
    C = 'C',
    Go = 'Go',
    Java = 'Java',
    Javascript = 'JavaScript',
    Php = 'PHP',
    Python = 'Python',
    Typescript = 'TypeScript',
}

export interface TourTaskType {
    title: string
    icon?: React.ReactNode
    steps: TourTaskStepType[]
    completed?: number
}

export interface TourTaskStepType {
    id: string
    label: string
    action: {
        type: 'video' | 'link' | 'restart'
        value: string | Record<TourLanguage, string>
    }
    info?: string
    completeAfterEvents?: string[]
    // TODO: add jsDocs
    isCompleted?: boolean
}
