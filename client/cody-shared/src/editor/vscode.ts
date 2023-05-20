import { ActiveTextEditorSelection } from '.'

export interface InlineController {
    selection: ActiveTextEditorSelection | null
}

export interface TaskContoller {
    add(input: string, selection: ActiveTextEditorSelection): string | null
    stop(taskID: string, content: string | null): Promise<void>
}
