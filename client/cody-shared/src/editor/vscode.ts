import { ActiveTextEditorSelection } from '.'

export interface InlineController {
    selection: ActiveTextEditorSelection | null
}

export interface TaskViewProvider {
    newTask(taskID: string, input: string, selection: ActiveTextEditorSelection, rootPath: string): void
    stopTask(taskID: string, content: string | null): Promise<void>
    refresh(): void
}
