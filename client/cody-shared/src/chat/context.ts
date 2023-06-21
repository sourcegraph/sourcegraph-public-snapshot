import { ConfigurationUseContext } from '../configuration'
import { ActiveTextEditorSelectionRange } from '../editor'

export interface ChatContextStatus {
    mode?: ConfigurationUseContext
    connection?: boolean
    codebase?: string
    filePath?: string
    selection?: ActiveTextEditorSelectionRange
    supportsKeyword?: boolean
}
