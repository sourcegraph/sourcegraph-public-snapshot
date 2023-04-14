import { CodebaseContext } from '../codebase-context'
import { Configuration } from '../configuration'
import { Editor } from '../editor'
import { SourcegraphEmbeddingsSearchClient } from '../embeddings/client'
import { SourcegraphIntentDetectorClient } from '../intent-detector/client'
import { KeywordContextFetcher } from '../keyword-context'
import { SourcegraphBrowserCompletionsClient } from '../sourcegraph-api/completions/browserClient'
import { SourcegraphGraphQLAPIClient } from '../sourcegraph-api/graphql/client'
import { isError } from '../utils'

import { BotResponseMultiplexer } from './bot-response-multiplexer'
import { ChatClient } from './chat'
import { ChatQuestion } from './recipes/chat-question'
import { Transcript } from './transcript'
import { ChatMessage } from './transcript/messages'
import { reformatBotMessage } from './viewHelpers'

export interface ClientInit {
    config: Pick<Configuration, 'serverEndpoint' | 'codebase' | 'useContext'>
    accessToken: string | null
    setMessageInProgress: (messageInProgress: ChatMessage | null) => void
    setTranscript: (transcript: ChatMessage[]) => void
    customHeaders?: Record<string, string>
}

export interface Client {
    submitMessage: (text: string) => void
}

export async function createClient({
    config,
    accessToken,
    setMessageInProgress,
    setTranscript,
    customHeaders,
}: ClientInit): Promise<Client> {
    const completionsClient = new SourcegraphBrowserCompletionsClient(
        config.serverEndpoint,
        accessToken,
        process.env.NODE_ENV === 'development' ? 'development' : 'production',
        customHeaders
    )
    const chatClient = new ChatClient(completionsClient)

    const graphqlClient = new SourcegraphGraphQLAPIClient(config.serverEndpoint, accessToken, customHeaders)

    const repoId = config.codebase ? await graphqlClient.getRepoId(config.codebase) : null
    if (isError(repoId)) {
        throw new Error(
            `Cody could not access the '${config.codebase}' repository on your Sourcegraph instance. Details: ${repoId.message}`
        )
    }

    const embeddingsSearch = repoId ? new SourcegraphEmbeddingsSearchClient(graphqlClient, repoId) : null

    const codebaseContext = new CodebaseContext(config.useContext, embeddingsSearch, noopKeywordFetcher)

    const intentDetector = new SourcegraphIntentDetectorClient(graphqlClient)

    const transcript = new Transcript()

    /* eslint-disable @typescript-eslint/require-await */
    const fakeEditor: Editor = {
        getActiveTextEditor() {
            return null
        },
        getActiveTextEditorSelection() {
            return null
        },
        getActiveTextEditorVisibleContent() {
            return null
        },
        getWorkspaceRootPath() {
            return null
        },
        replaceSelection(_fileName, _selectedText, _replacement) {
            return Promise.resolve()
        },
        async showQuickPick(labels) {
            return window.prompt(`Choose: ${labels.join(', ')}`, labels[0]) || undefined
        },
        async showWarningMessage(message) {
            console.warn(message)
        },
    }
    /* eslint-enable @typescript-eslint/require-await */

    let isMessageInProgress = false

    const sendTranscript = (): void => {
        if (isMessageInProgress) {
            const messages = transcript.toChat()
            setTranscript(messages.slice(0, -1))
            setMessageInProgress(messages[messages.length - 1])
        } else {
            setTranscript(transcript.toChat())
            setMessageInProgress(null)
        }
    }

    const chatQuestionRecipe = new ChatQuestion()

    return {
        submitMessage: async (text: string) => {
            const interaction = await chatQuestionRecipe.getInteraction(text, {
                editor: fakeEditor,
                intentDetector,
                codebaseContext,
                responseMultiplexer: new BotResponseMultiplexer(),
            })
            if (!interaction) {
                throw new Error('No interaction')
            }
            isMessageInProgress = true
            transcript.addInteraction(interaction)
            sendTranscript()

            const prompt = await transcript.toPrompt()
            const responsePrefix = interaction.getAssistantMessage().prefix ?? ''

            chatClient.chat(prompt, {
                onChange(rawText) {
                    const text = reformatBotMessage(rawText, responsePrefix)
                    transcript.addAssistantResponse(text)

                    sendTranscript()
                },
                onComplete() {
                    isMessageInProgress = false
                    sendTranscript()
                },
                onError(error) {
                    console.error(error)
                },
            })
        },
    }
}

const noopKeywordFetcher: KeywordContextFetcher = {
    // eslint-disable-next-line @typescript-eslint/require-await
    async getContext() {
        throw new Error('noopKeywordFetcher: not implemented')
    },
}
