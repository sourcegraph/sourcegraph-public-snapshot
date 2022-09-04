import { ObservableInput } from 'rxjs'

import { Settings } from '../app/useSettings'

export interface TaskProvider {
    name: string
    tasks: (settings: Settings) => ObservableInput<Task[]>
}

export interface Task {
    text: string
    url?: string
}
