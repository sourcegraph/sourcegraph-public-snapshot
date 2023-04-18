#! /usr/bin/env node
import { Command } from 'commander'
import { cleanEnv, str } from 'envalid'
import { memoize } from 'lodash'

import { Transcript } from '@sourcegraph/cody-shared/src/chat/transcript'
import { Interaction } from '@sourcegraph/cody-shared/src/chat/transcript/interaction'
import { CodebaseContext } from '@sourcegraph/cody-shared/src/codebase-context'
import { ContextMessage } from '@sourcegraph/cody-shared/src/codebase-context/messages'
import { SourcegraphEmbeddingsSearchClient } from '@sourcegraph/cody-shared/src/embeddings/client'
import { IntentDetector } from '@sourcegraph/cody-shared/src/intent-detector'
import { SourcegraphIntentDetectorClient } from '@sourcegraph/cody-shared/src/intent-detector/client'
import { KeywordContextFetcher } from '@sourcegraph/cody-shared/src/keyword-context'
import { SOLUTION_TOKEN_LENGTH, MAX_HUMAN_INPUT_TOKENS } from '@sourcegraph/cody-shared/src/prompt/constants'
import { truncateText } from '@sourcegraph/cody-shared/src/prompt/truncation'
import { SourcegraphNodeCompletionsClient } from '@sourcegraph/cody-shared/src/sourcegraph-api/completions/nodeClient'
import {
    CompletionParameters,
    CompletionCallbacks,
    Message,
} from '@sourcegraph/cody-shared/src/sourcegraph-api/completions/types'
import { SourcegraphGraphQLAPIClient } from '@sourcegraph/cody-shared/src/sourcegraph-api/graphql'
import { isError } from '@sourcegraph/cody-shared/src/utils'

import { getPreamble } from './preamble'

const ENVIRONMENT_CONFIG = cleanEnv(process.env, {
    SRC_ACCESS_TOKEN: str(),
})

const DEFAULT_APP_SETTINGS = {
    codebase: 'github.com/sourcegraph/sourcegraph',
    serverEndpoint: 'https://sourcegraph.sourcegraph.com',
    contextType: 'blended',
    debug: 'development',
} as const

/**
 * Memoized function to get the repository ID for a given codebase.
 */
const getRepoId = memoize(async (client: SourcegraphGraphQLAPIClient, codebase: string) => {
    const repoId = codebase ? await client.getRepoId(codebase) : null

    if (isError(repoId)) {
        const errorMessage =
            `Cody could not find the '${codebase}' repository on your Sourcegraph instance.\n` +
            'Please check that the repository exists and is entered correctly in the cody.codebase setting.'
        console.error(errorMessage)
    }

    return repoId
})

export async function createCodebaseContext(
    client: SourcegraphGraphQLAPIClient,
    codebase: string,
    contextType: 'embeddings' | 'keyword' | 'none' | 'blended'
) {
    const repoId = await getRepoId(client, codebase)
    const embeddingsSearch = repoId && !isError(repoId) ? new SourcegraphEmbeddingsSearchClient(client, repoId) : null

    const codebaseContext = new CodebaseContext(
        { useContext: contextType },
        embeddingsSearch,
        new LocalKeywordContextFetcherMock()
    )

    return codebaseContext
}

class LocalKeywordContextFetcherMock implements KeywordContextFetcher {
    public getContext() {
        return Promise.resolve([])
    }
}

const DEFAULT_CHAT_COMPLETION_PARAMETERS: Omit<CompletionParameters, 'messages'> = {
    temperature: 0.2,
    maxTokensToSample: SOLUTION_TOKEN_LENGTH,
    topK: -1,
    topP: -1,
}

function streamCompletions(client: SourcegraphNodeCompletionsClient, messages: Message[], cb: CompletionCallbacks) {
    return client.stream({ messages, ...DEFAULT_CHAT_COMPLETION_PARAMETERS }, cb)
}

async function getContextMessages(
    text: string,
    intentDetector: IntentDetector,
    codebaseContext: CodebaseContext
): Promise<ContextMessage[]> {
    const contextMessages: ContextMessage[] = []

    const isCodebaseContextRequired = await intentDetector.isCodebaseContextRequired(text)

    if (isCodebaseContextRequired) {
        const codebaseContextMessages = await codebaseContext.getContextMessages(text, {
            numCodeResults: 8,
            numTextResults: 2,
        })

        contextMessages.push(...codebaseContextMessages)
    }

    return contextMessages
}

async function interactionFromMessage(
    message: Message,
    intentDetector: IntentDetector,
    codebaseContext: CodebaseContext | null
): Promise<Interaction | null> {
    if (!message.text) {
        return Promise.resolve(null)
    }

    const text = truncateText(message.text, MAX_HUMAN_INPUT_TOKENS)

    const contextMessages =
        codebaseContext === null ? Promise.resolve([]) : getContextMessages(text, intentDetector, codebaseContext)

    return Promise.resolve(
        new Interaction(
            { speaker: 'human', text, displayText: text },
            { speaker: 'assistant', text: '', displayText: '' },
            contextMessages
        )
    )
}

async function startCLI() {
    const program = new Command()

    program
        .version('0.0.1')
        .description('Cody CLI')
        .option('-p, --prompt <value>', 'Give Cody a prompt')
        .parse(process.argv)

    const options = program.opts()
    if (options.prompt === '' || options.prompt === undefined) {
        console.error('no prompt with --prompt provided')
        process.exit(1)
    }

    const sourcegraphClient = new SourcegraphGraphQLAPIClient({
        serverEndpoint: DEFAULT_APP_SETTINGS.serverEndpoint,
        accessToken: ENVIRONMENT_CONFIG.SRC_ACCESS_TOKEN,
        customHeaders: {},
    })

    const intentDetector = new SourcegraphIntentDetectorClient(sourcegraphClient)

    const codebaseContext = await createCodebaseContext(
        sourcegraphClient,
        DEFAULT_APP_SETTINGS.codebase,
        DEFAULT_APP_SETTINGS.contextType
    )

    const completionsClient = new SourcegraphNodeCompletionsClient({
        serverEndpoint: DEFAULT_APP_SETTINGS.serverEndpoint,
        accessToken: ENVIRONMENT_CONFIG.SRC_ACCESS_TOKEN,
        debug: DEFAULT_APP_SETTINGS.debug === 'development',
        customHeaders: {},
    })

    const transcript = new Transcript()

    // TODO: Keep track of all user input if we add REPL mode

    const initialMessage: Message = { speaker: 'human', text: options.prompt }
    const messages: { human: Message; assistant?: Message }[] = [{ human: initialMessage }]
    for (const [index, message] of messages.entries()) {
        const interaction = await interactionFromMessage(
            message.human,
            intentDetector,
            // Fetch codebase context only for the last message
            index === messages.length - 1 ? codebaseContext : null
        )

        transcript.addInteraction(interaction)

        if (message.assistant?.text) {
            transcript.addAssistantResponse(message.assistant?.text)
        }
    }

    const prompt = await transcript.toPrompt(getPreamble(DEFAULT_APP_SETTINGS.codebase))

    let text = ''
    streamCompletions(completionsClient, prompt, {
        onChange: chunk => {
            text = chunk
        },
        onComplete: () => {
            console.log(text)
        },
        onError: err => {
            console.error(err)
        },
    })
}

startCLI()
    .then(() => {})
    .catch(error => {
        console.error('Error starting the bot:', error)
        process.exit(1)
    })
