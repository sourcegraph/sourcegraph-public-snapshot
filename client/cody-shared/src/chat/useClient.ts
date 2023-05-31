import { useState, useCallback, useMemo } from 'react'

import { CodebaseContext } from '../codebase-context'
import { ConfigurationWithAccessToken } from '../configuration'
import { Editor, NoopEditor } from '../editor'
import { PrefilledOptions, withPreselectedOptions } from '../editor/withPreselectedOptions'
import { SourcegraphEmbeddingsSearchClient } from '../embeddings/client'
import { SourcegraphIntentDetectorClient } from '../intent-detector/client'
import { SourcegraphBrowserCompletionsClient } from '../sourcegraph-api/completions/browserClient'
import { SourcegraphGraphQLAPIClient } from '../sourcegraph-api/graphql'
import { isError } from '../utils'

import { BotResponseMultiplexer } from './bot-response-multiplexer'
import { ChatClient } from './chat'
import { ChatContextStatus } from './context'
import { getPreamble } from './preamble'
import { getRecipe } from './recipes/browser-recipes'
import { RecipeID } from './recipes/recipe'
import { Transcript } from './transcript'
import { ChatMessage } from './transcript/messages'
import { reformatBotMessage } from './viewHelpers'

export type CodyClientConfig = Pick<
    ConfigurationWithAccessToken,
    'serverEndpoint' | 'useContext' | 'accessToken' | 'customHeaders'
> & { debugEnable: boolean; needsEmailVerification: boolean }

export interface CodyClientScope {
    type: 'Automatic' | 'None' | 'Repositories'
    repositories: string[]
    editor: Editor
}

export type CodyClientEvent = 'submit' | 'initializedNewChat' | 'error'

export interface CodyClient {
    readonly transcript: Transcript | null
    readonly chatMessages: ChatMessage[]
    readonly messageInProgress: ChatMessage | null
    readonly isMessageInProgress: boolean
    readonly scope: CodyClientScope
    readonly config: CodyClientConfig
    readonly legacyChatContext: ChatContextStatus
    setTranscript: (transcript: Transcript) => Promise<void>
    setScope: (scope: CodyClientScope) => void
    setConfig: (config: CodyClientConfig) => void
    submitMessage: (humanChatInput: string, scope?: CodyClientScope) => Promise<Transcript | null>
    editMessage: (
        humanChatInput: string,
        messageId?: string | undefined,
        scope?: CodyClientScope
    ) => Promise<Transcript | null>
    initializeNewChat: () => Transcript | null
    executeRecipe: (
        recipeId: RecipeID,
        options?: {
            prefilledOptions?: PrefilledOptions
            humanChatInput?: string
            scope?: CodyClientScope
        }
    ) => Promise<Transcript | null>
    setEditorScope: (editor: Editor) => void
}

interface CodyClientProps {
    config: CodyClientConfig
    scope?: CodyClientScope
    initialTranscript?: Transcript | null
    onEvent?: (event: CodyClientEvent) => void
    web?: boolean
}

export const useClient = ({
    config: initialConfig,
    initialTranscript = null,
    scope: initialScope = {
        type: 'None',
        repositories: [],
        editor: new NoopEditor(),
    },
    onEvent,
    web = false,
}: CodyClientProps): CodyClient => {
    const [transcript, setTranscriptState] = useState<Transcript | null>(initialTranscript)
    const [chatMessages, setChatMessagesState] = useState<ChatMessage[]>([])
    const [isMessageInProgress, setIsMessageInProgressState] = useState<boolean>(false)

    const messageInProgress: ChatMessage | null = useMemo(() => {
        if (isMessageInProgress) {
            const lastMessage = chatMessages[chatMessages.length - 1]

            if (lastMessage?.speaker === 'assistant') {
                return lastMessage
            }
        }

        return null
    }, [chatMessages, isMessageInProgress])

    const setTranscript = useCallback(async (transcript: Transcript): Promise<void> => {
        const messages = await transcript.toChatPromise()

        setIsMessageInProgressState(false)
        setTranscriptState(transcript)
        setChatMessagesState(messages)
    }, [])

    const [config, setConfig] = useState<CodyClientConfig>(initialConfig)

    const initializeNewChat = useCallback((): Transcript | null => {
        if (config.needsEmailVerification) {
            return transcript
        }
        const newTranscript = new Transcript()
        setIsMessageInProgressState(false)
        setTranscriptState(newTranscript)
        setChatMessagesState(newTranscript.toChat())
        onEvent?.('initializedNewChat')

        return newTranscript
    }, [onEvent, config.needsEmailVerification, transcript])

    const { graphqlClient, chatClient, intentDetector } = useMemo(() => {
        const completionsClient = new SourcegraphBrowserCompletionsClient(config)
        const chatClient = new ChatClient(completionsClient)
        const graphqlClient = new SourcegraphGraphQLAPIClient(config)
        const intentDetector = new SourcegraphIntentDetectorClient(graphqlClient)

        return { graphqlClient, chatClient, intentDetector }
    }, [config])

    const [scope, setScopeState] = useState<CodyClientScope>(initialScope)
    const setScope = useCallback((scope: CodyClientScope) => {
        setScopeState(scope)
    }, [])
    const setEditorScope = useCallback((editor: Editor) => {
        setScopeState(scope => ({ ...scope, editor }))
    }, [])

    // TODO(naman): temporarily set codebase to the first repository in the list until multi-repo context is implemented throughout.
    const codebase: string | null = useMemo(() => scope.repositories[0] || null, [scope])
    const codebaseId: Promise<string | null> = useMemo(async () => {
        if (codebase === null) {
            return null
        }

        const id = (await graphqlClient.getRepoIdIfEmbeddingExists(codebase)) || null
        if (isError(id)) {
            console.error(
                `Cody could not access the '${codebase}' repository on your Sourcegraph instance. Details: ${id.message}`
            )
            return null
        }

        return id
    }, [codebase, graphqlClient])

    const executeRecipe = useCallback(
        async (
            recipeId: RecipeID,
            options?: {
                prefilledOptions?: PrefilledOptions
                humanChatInput?: string
                // TODO(naman): accept scope with execute recipe
                scope?: CodyClientScope
            }
        ): Promise<Transcript | null> => {
            const recipe = getRecipe(recipeId)
            if (!recipe || transcript === null || isMessageInProgress || config.needsEmailVerification) {
                return Promise.resolve(null)
            }

            const repoId = await codebaseId
            const embeddingsSearch = repoId ? new SourcegraphEmbeddingsSearchClient(graphqlClient, repoId, web) : null
            const codebaseContext = new CodebaseContext(config, codebase || undefined, embeddingsSearch, null)

            const { humanChatInput = '', prefilledOptions } = options ?? {}
            // TODO(naman): save scope with each interaction
            const interaction = await recipe.getInteraction(humanChatInput, {
                editor: prefilledOptions ? withPreselectedOptions(scope.editor, prefilledOptions) : scope.editor,
                intentDetector,
                codebaseContext,
                responseMultiplexer: new BotResponseMultiplexer(),
                firstInteraction: transcript.isEmpty,
            })
            if (!interaction) {
                return Promise.resolve(null)
            }

            transcript.addInteraction(interaction)
            setChatMessagesState(transcript.toChat())
            setIsMessageInProgressState(true)
            onEvent?.('submit')

            const prompt = await transcript.toPrompt(getPreamble(codebase || undefined))
            const responsePrefix = interaction.getAssistantMessage().prefix ?? ''
            let rawText = ''

            return new Promise(resolve => {
                chatClient.chat(prompt, {
                    onChange(_rawText) {
                        rawText = _rawText

                        const text = reformatBotMessage(rawText, responsePrefix)
                        transcript.addAssistantResponse(text)
                        setChatMessagesState(transcript.toChat())
                    },
                    onComplete() {
                        const text = reformatBotMessage(rawText, responsePrefix)
                        transcript.addAssistantResponse(text)

                        transcript
                            .toChatPromise()
                            .then(messages => {
                                setChatMessagesState(messages)
                                setIsMessageInProgressState(false)
                            })
                            .catch(() => null)

                        resolve(transcript)
                    },
                    onError(error) {
                        // Display error message as assistant response
                        transcript.addErrorAsAssistantResponse(
                            `<div class="cody-chat-error"><span>Request failed: </span>${error}</div>`
                        )

                        console.error(`Completion request failed: ${error}`)

                        transcript
                            .toChatPromise()
                            .then(messages => {
                                setChatMessagesState(messages)
                                setIsMessageInProgressState(false)
                            })
                            .catch(() => null)

                        onEvent?.('error')
                        resolve(transcript)
                    },
                })
            })
        },
        [
            config,
            scope,
            codebase,
            codebaseId,
            graphqlClient,
            transcript,
            intentDetector,
            chatClient,
            isMessageInProgress,
            onEvent,
            web,
        ]
    )

    const submitMessage = useCallback(
        async (humanChatInput: string, scope?: CodyClientScope): Promise<Transcript | null> =>
            executeRecipe('chat-question', { humanChatInput, scope }),
        [executeRecipe]
    )

    // TODO(naman): load message scope from the interaction
    const editMessage = useCallback(
        async (
            humanChatInput: string,
            messageId?: string | undefined,
            scope?: CodyClientScope
        ): Promise<Transcript | null> => {
            if (!transcript) {
                return transcript
            }

            const timestamp = messageId || transcript.getLastInteraction()?.timestamp || new Date().toISOString()

            transcript.removeInteractionsSince(timestamp)
            setChatMessagesState(transcript.toChat())

            return submitMessage(humanChatInput, scope)
        },
        [transcript, submitMessage]
    )

    // TODO(naman): usage of `chatContext` in Chat UI component will be replaced by `scope`. Remove this once done.
    const legacyChatContext = useMemo<ChatContextStatus>(
        () => ({
            codebase: codebase || undefined,
            filePath: scope.editor.getActiveTextEditorSelectionOrEntireFile()?.fileName,
            connection: true,
        }),
        [codebase, scope]
    )

    const returningChatMessages = useMemo(
        () => (messageInProgress ? chatMessages.slice(0, -1) : chatMessages),
        [chatMessages, messageInProgress]
    )

    return useMemo(
        () => ({
            transcript,
            chatMessages: returningChatMessages,
            isMessageInProgress,
            messageInProgress,
            setTranscript,
            scope,
            setScope,
            setEditorScope,
            config,
            setConfig,
            executeRecipe,
            submitMessage,
            initializeNewChat,
            editMessage,
            legacyChatContext,
        }),
        [
            transcript,
            returningChatMessages,
            isMessageInProgress,
            messageInProgress,
            setTranscript,
            scope,
            setScope,
            setEditorScope,
            config,
            setConfig,
            executeRecipe,
            submitMessage,
            initializeNewChat,
            editMessage,
            legacyChatContext,
        ]
    )
}
