#! /usr/bin/env node
import { Command } from 'commander'
import prompts from 'prompts'

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
import { isRepoNotFoundError } from '@sourcegraph/cody-shared/src/sourcegraph-api/graphql/client'
import { isError } from '@sourcegraph/cody-shared/src/utils'

import { DEFAULTS, ENVIRONMENT_CONFIG } from './config'
import { getPreamble } from './preamble'

/**
 * Memoized function to get the repository ID for a given codebase.
 */
const getRepoId = async (client: SourcegraphGraphQLAPIClient, codebase: string) => {
    const repoId = codebase ? await client.getRepoId(codebase) : null
    return repoId
}

export async function createCodebaseContext(
    client: SourcegraphGraphQLAPIClient,
    codebase: string,
    contextType: 'embeddings' | 'keyword' | 'none' | 'blended'
) {
    const repoId = await getRepoId(client, codebase)
    if (isError(repoId)) {
        throw repoId
    }

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
        .option('-c, --codebase <value>', `Codebase to use for context fetching. Default: ${DEFAULTS.codebase}`)
        .option('-e, --endpoint <value>', `Sourcegraph instance to connect to. Default: ${DEFAULTS.serverEndpoint}`)
        .option(
            '--context [embeddings,keyword,none,blended]',
            `How Cody fetches context for query. Default: ${DEFAULTS.contextType}`
        )
        .parse(process.argv)

    const options = program.opts()

    const codebase: string = options.codebase || DEFAULTS.codebase
    const endpoint: string = options.endpoint || DEFAULTS.serverEndpoint
    const contextType: 'keyword' | 'embeddings' | 'none' | 'blended' = options.contextType || DEFAULTS.contextType
    const accessToken: string | undefined = ENVIRONMENT_CONFIG.SRC_ACCESS_TOKEN
    if (accessToken === undefined || accessToken === '') {
        console.error(
            `No access token found. Set SRC_ACCESS_TOKEN to an access token created on the Sourcegraph instance.`
        )
        process.exit(1)
    }

    const sourcegraphClient = new SourcegraphGraphQLAPIClient({
        serverEndpoint: endpoint,
        accessToken: accessToken,
        customHeaders: {},
    })

    let codebaseContext
    try {
        codebaseContext = await createCodebaseContext(sourcegraphClient, codebase, contextType)
    } catch (error) {
        let errorMessage = ''
        if (isRepoNotFoundError(error)) {
            errorMessage =
                `Cody could not find the '${codebase}' repository on your Sourcegraph instance.\n` +
                'Please check that the repository exists and is entered correctly in the cody.codebase setting.'
        } else {
            errorMessage =
                `Cody could not connect to your Sourcegraph instance: ${error}\n` +
                'Make sure that cody.serverEndpoint is set to a running Sourcegraph instance and that an access token is configured.'
        }
        console.error(errorMessage)
        process.exit(1)
    }

    const intentDetector = new SourcegraphIntentDetectorClient(sourcegraphClient)

    const completionsClient = new SourcegraphNodeCompletionsClient({
        serverEndpoint: endpoint,
        accessToken: ENVIRONMENT_CONFIG.SRC_ACCESS_TOKEN,
        debug: DEFAULTS.debug === 'development',
        customHeaders: {},
    })

    let prompt = options.prompt
    if (prompt === undefined || prompt === '') {
        const response = await prompts({
            type: 'text',
            name: 'value',
            message: 'What do you want to ask Cody?',
        })

        prompt = response.value
    }

    const transcript = new Transcript()

    // TODO: Keep track of all user input if we add REPL mode

    const initialMessage: Message = { speaker: 'human', text: prompt }
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

    const finalPrompt = await transcript.toPrompt(getPreamble(codebase))

    let text = ''
    streamCompletions(completionsClient, finalPrompt, {
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
