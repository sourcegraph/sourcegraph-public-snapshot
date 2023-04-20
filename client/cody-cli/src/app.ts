#! /usr/bin/env node
import { Command } from 'commander'
import prompts from 'prompts'

import { Transcript } from '@sourcegraph/cody-shared/src/chat/transcript'
import { SourcegraphIntentDetectorClient } from '@sourcegraph/cody-shared/src/intent-detector/client'
import { SourcegraphNodeCompletionsClient } from '@sourcegraph/cody-shared/src/sourcegraph-api/completions/nodeClient'
import { Message } from '@sourcegraph/cody-shared/src/sourcegraph-api/completions/types'
import { SourcegraphGraphQLAPIClient } from '@sourcegraph/cody-shared/src/sourcegraph-api/graphql'
import { isRepoNotFoundError } from '@sourcegraph/cody-shared/src/sourcegraph-api/graphql/client'

import { streamCompletions } from './completions'
import { DEFAULTS, ENVIRONMENT_CONFIG } from './config'
import { createCodebaseContext } from './context'
import { interactionFromMessage } from './interactions'
import { getPreamble } from './preamble'

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
        .option('--lsp', 'Start LSP')
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
