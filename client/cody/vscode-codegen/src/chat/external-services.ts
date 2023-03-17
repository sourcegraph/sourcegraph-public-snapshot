import { getAccessToken, SecretStorage } from '../command/secret-storage'
import { Embeddings } from '../embeddings'
import { EmbeddingsClient } from '../embeddings/client'
import { LLMIntentDetector } from '../intent-detector/llm-intent-detector'
import { SourcegraphCompletionsClient } from '../sourcegraph-api/completions'
import { SourcegraphGraphQLAPIClient } from '../sourcegraph-api/graphql'
import { isError } from '../utils'

import { ChatClient } from './chat'

interface ExternalServices {
    intentDetector: LLMIntentDetector
    embeddings: Embeddings | null
    chatClient: ChatClient
}

export async function configureExternalServices(
    codebase: string,
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

    return {
        intentDetector: new LLMIntentDetector(completions),
        embeddings: repoId && !isError(repoId) ? new EmbeddingsClient(client, repoId) : null,
        chatClient: new ChatClient(completions),
    }
}
