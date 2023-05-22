import { ConfigurationUseContext } from '../configuration'

export interface ChatContextStatus {
    mode?: ConfigurationUseContext
    connection?: boolean
    codebase?: string
    filePath?: string
    supportsKeyword?: boolean
}
