import { ChatClient } from './chat/chat'
import { CodebaseContext } from './codebase-context'
import { getAccessToken, SecretStorage } from './command/secret-storage'
import { SourcegraphEmbeddingsSearchClient } from './embeddings/client'
import { IntentDetector } from './intent-detector'
import { SourcegraphIntentDetectorClient } from './intent-detector/client'
import { LocalKeywordContextFetcher } from './keyword-context/local-keyword-context-fetcher'
import { SourcegraphCompletionsClient } from './sourcegraph-api/completions'
import { SourcegraphGraphQLAPIClient } from './sourcegraph-api/graphql'
import { isError } from './utils'

interface ExternalServices {
    intentDetector: IntentDetector
    codebaseContext: CodebaseContext
    chatClient: ChatClient
}

export async function configureExternalServices(
    contextType: 'embeddings' | 'keyword' | 'none' | 'blended',
    codebase: string,
    rgPath: string,
    serverEndpoint: string,
    secretStorage: SecretStorage,
    mode: 'development' | 'production'
): Promise<ExternalServices> {
    const accessToken = await getAccessToken(secretStorage)
    const client = new SourcegraphGraphQLAPIClient(serverEndpoint, accessToken)
    const completions = new SourcegraphCompletionsClient(serverEndpoint, accessToken, mode)

    const repoId = codebase ? await client.getRepoId(codebase) : null
    if (isError(repoId)) {
        console.error('error fetching codebase', codebase)
    }
    const embeddingsSearch = repoId && !isError(repoId) ? new SourcegraphEmbeddingsSearchClient(client, repoId) : null

    const codebaseContext = new CodebaseContext(contextType, embeddingsSearch, new LocalKeywordContextFetcher(rgPath))

    return {
        intentDetector: new SourcegraphIntentDetectorClient(client),
        codebaseContext,
        chatClient: new ChatClient(completions),
    }
}
