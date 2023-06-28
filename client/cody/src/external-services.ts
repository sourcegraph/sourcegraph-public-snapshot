import { ChatClient } from '@sourcegraph/cody-shared/src/chat/chat'
import { CodebaseContext } from '@sourcegraph/cody-shared/src/codebase-context'
import { ConfigurationWithAccessToken } from '@sourcegraph/cody-shared/src/configuration'
import { Editor } from '@sourcegraph/cody-shared/src/editor'
import { SourcegraphEmbeddingsSearchClient } from '@sourcegraph/cody-shared/src/embeddings/client'
import { Guardrails } from '@sourcegraph/cody-shared/src/guardrails'
import { SourcegraphGuardrailsClient } from '@sourcegraph/cody-shared/src/guardrails/client'
import { IntentDetector } from '@sourcegraph/cody-shared/src/intent-detector'
import { SourcegraphIntentDetectorClient } from '@sourcegraph/cody-shared/src/intent-detector/client'
import { SourcegraphCompletionsClient } from '@sourcegraph/cody-shared/src/sourcegraph-api/completions/client'
import { SourcegraphNodeCompletionsClient } from '@sourcegraph/cody-shared/src/sourcegraph-api/completions/nodeClient'
import { SourcegraphGraphQLAPIClient } from '@sourcegraph/cody-shared/src/sourcegraph-api/graphql'
import { isError } from '@sourcegraph/cody-shared/src/utils'

import { VSCodeGraphContextFetcher } from './graph-context/graph-context-fetcher'
import { FilenameContextFetcher } from './local-context/filename-context-fetcher'
import { LocalKeywordContextFetcher } from './local-context/local-keyword-context-fetcher'
import { logger } from './log'
import { getRerankWithLog } from './logged-rerank'

interface ExternalServices {
    intentDetector: IntentDetector
    codebaseContext: CodebaseContext
    chatClient: ChatClient
    completionsClient: SourcegraphCompletionsClient
    guardrails: Guardrails

    /** Update configuration for all of the services in this interface. */
    onConfigurationChange: (newConfig: ExternalServicesConfiguration) => void
}

type ExternalServicesConfiguration = Pick<
    ConfigurationWithAccessToken,
    'serverEndpoint' | 'codebase' | 'useContext' | 'customHeaders' | 'accessToken' | 'debugEnable'
>

export async function configureExternalServices(
    initialConfig: ExternalServicesConfiguration,
    rgPath: string,
    editor: Editor
): Promise<ExternalServices> {
    const client = new SourcegraphGraphQLAPIClient(initialConfig)
    const completions = new SourcegraphNodeCompletionsClient(initialConfig, logger)

    const repoId = initialConfig.codebase ? await client.getRepoId(initialConfig.codebase) : null
    if (isError(repoId)) {
        const infoMessage =
            `Cody could not find the '${initialConfig.codebase}' repository on your Sourcegraph instance.\n` +
            'Please check that the repository exists. You can override the repository with the "cody.codebase" setting.'
        console.info(infoMessage)
    }
    const embeddingsSearch = repoId && !isError(repoId) ? new SourcegraphEmbeddingsSearchClient(client, repoId) : null

    const chatClient = new ChatClient(completions)
    const codebaseContext = new CodebaseContext(
        initialConfig,
        initialConfig.codebase,
        embeddingsSearch,
        new LocalKeywordContextFetcher(rgPath, editor, chatClient),
        new FilenameContextFetcher(rgPath, editor, chatClient),
        new VSCodeGraphContextFetcher(client, editor),
        undefined,
        getRerankWithLog(chatClient)
    )

    const guardrails = new SourcegraphGuardrailsClient(client)

    return {
        intentDetector: new SourcegraphIntentDetectorClient(client),
        codebaseContext,
        chatClient,
        completionsClient: completions,
        guardrails,
        onConfigurationChange: newConfig => {
            client.onConfigurationChange(newConfig)
            completions.onConfigurationChange(newConfig)
            codebaseContext.onConfigurationChange(newConfig)
        },
    }
}
