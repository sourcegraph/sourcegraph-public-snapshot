import { ConfigurationUseContext } from '../configuration'
import { ActiveTextEditorSelectionRange } from '../editor'
import { RepoEmbeddingJobState, RepoEmbeddingJobsConnection } from '../sourcegraph-api/graphql/client'

export interface ChatContextStatus {
    mode?: ConfigurationUseContext
    connection?: boolean
    codebase?: string
    filePath?: string
    selection?: ActiveTextEditorSelectionRange
    supportsKeyword?: boolean
    repoStatus?: RepoEmbeddingJobState
    indexStatus?: RepoEmbeddingJobsConnection
    isApp?: boolean
}
