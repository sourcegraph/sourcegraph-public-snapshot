import { useState, useCallback, useMemo } from 'react'

import { CodebaseContext } from '../codebase-context'
import { ConfigurationWithAccessToken } from '../configuration'
import { Editor, NoopEditor } from '../editor'
import { PrefilledOptions, withPreselectedOptions } from '../editor/withPreselectedOptions'
import { SourcegraphIntentDetectorClient } from '../intent-detector/client'
import { SourcegraphBrowserCompletionsClient } from '../sourcegraph-api/completions/browserClient'
import { SourcegraphGraphQLAPIClient } from '../sourcegraph-api/graphql'
import { UnifiedContextFetcherClient } from '../unified-context/client'
import { isError } from '../utils'

import { BotResponseMultiplexer } from './bot-response-multiplexer'
import { ChatClient } from './chat'
import { getMultiRepoPreamble } from './preamble'
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
    includeInferredRepository: boolean
    includeInferredFile: boolean
    repositories: string[]
    editor: Editor
}

export interface CodyClientScopePartial {
    repositories?: string[]
    editor?: Editor
}

export type CodyClientEvent = 'submit' | 'initializedNewChat' | 'error'

export interface CodyClient {
    readonly transcript: Transcript | null
    readonly chatMessages: ChatMessage[]
    readonly messageInProgress: ChatMessage | null
    readonly isMessageInProgress: boolean
    readonly scope: CodyClientScope
    readonly config: CodyClientConfig
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
            scope?: {
                editor?: Editor
            }
        }
    ) => Promise<Transcript | null>
    setEditorScope: (editor: Editor) => void
    toggleIncludeInferredRepository: () => void
    toggleIncludeInferredFile: () => void
    abortMessageInProgress: () => void
}

interface CodyClientProps {
    config: CodyClientConfig
    scope?: CodyClientScope
    initialTranscript?: Transcript | null
    onEvent?: (event: CodyClientEvent) => void
}

export const useClient = ({
    config: initialConfig,
    initialTranscript = null,
    scope: initialScope = {
        includeInferredRepository: true,
        includeInferredFile: true,
        repositories: [],
        editor: new NoopEditor(),
    },
    onEvent,
}: CodyClientProps): CodyClient => {
    const [transcript, setTranscriptState] = useState<Transcript | null>(initialTranscript)
    const [chatMessages, setChatMessagesState] = useState<ChatMessage[]>([])
    const [isMessageInProgress, setIsMessageInProgressState] = useState<boolean>(false)
    const [abortMessageInProgressInternal, setAbortMessageInProgress] = useState<() => void>(() => () => undefined)

    const messageInProgress: ChatMessage | null = useMemo(() => {
        if (isMessageInProgress) {
            const lastMessage = chatMessages[chatMessages.length - 1]

            if (lastMessage?.speaker === 'assistant') {
                return lastMessage
            }
        }

        return null
    }, [chatMessages, isMessageInProgress])

    const abortMessageInProgress = useCallback(() => {
        abortMessageInProgressInternal()

        transcript
            ?.toChatPromise()
            .then(messages => {
                setChatMessagesState(messages)
                setIsMessageInProgressState(false)
            })
            .catch(error => console.error(`aborting in progress message failed: ${error}`))
    }, [abortMessageInProgressInternal, transcript, setChatMessagesState, setIsMessageInProgressState])

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
    const setScope = useCallback((scope: CodyClientScope) => setScopeState(scope), [setScopeState])

    const setEditorScope = useCallback(
        (editor: Editor) => setScopeState(scope => ({ ...scope, editor })),
        [setScopeState]
    )

    const toggleIncludeInferredRepository = useCallback(
        () =>
            setScopeState(scope => ({
                ...scope,
                includeInferredRepository: !scope.includeInferredRepository,
                includeInferredFile: !scope.includeInferredRepository,
            })),
        [setScopeState]
    )

    const toggleIncludeInferredFile = useCallback(
        () => setScopeState(scope => ({ ...scope, includeInferredFile: !scope.includeInferredFile })),
        [setScopeState]
    )

    const activeEditor = useMemo(() => scope.editor.getActiveTextEditor(), [scope.editor])

    const codebases: string[] = useMemo(() => {
        const repos = [...scope.repositories]
        if (scope.includeInferredRepository && activeEditor?.repoName) {
            repos.push(activeEditor.repoName)
        }

        return repos
    }, [scope, activeEditor])

    const codebaseIds: Promise<string[]> = useMemo(async () => {
        if (!codebases.length) {
            return []
        }

        const results = await graphqlClient.getRepoIds(codebases)
        if (isError(results)) {
            console.error(
                `Cody could not access the repositories on your Sourcegraph instance. Details: ${results.message}`
            )
            return []
        }

        return results.map(({ id }) => id)
    }, [codebases, graphqlClient])

    const executeRecipe = useCallback(
        async (
            recipeId: RecipeID,
            options?: {
                prefilledOptions?: PrefilledOptions
                humanChatInput?: string
                scope?: CodyClientScopePartial
            }
        ): Promise<Transcript | null> => {
            const recipe = getRecipe(recipeId)
            if (!recipe || transcript === null || isMessageInProgress || config.needsEmailVerification) {
                return Promise.resolve(null)
            }

            const repoNames = [...codebases]
            const repoIds = [...(await codebaseIds)]
            const editor = options?.scope?.editor || (scope.includeInferredFile ? scope.editor : new NoopEditor())
            const activeEditor = editor.getActiveTextEditor()
            if (activeEditor?.repoName && !repoNames.includes(activeEditor.repoName)) {
                // NOTE(naman): We allow users to disable automatic inferrence of current file & repo
                // using `includeInferredFile` and `includeInferredRepository` options. But for editor recipes
                // like "Explain code at high level", we need to pass the current repo & file context.
                // Here we are passing the current repo & file context based on `options.scope.editor`
                // if present.
                const additionalRepoId = await graphqlClient.getRepoId(activeEditor.repoName)
                if (isError(additionalRepoId)) {
                    console.error(
                        `Cody could not access the ${activeEditor.repoName} repository on your Sourcegraph instance. Details: ${additionalRepoId.message}`
                    )
                } else {
                    repoIds.push(additionalRepoId)
                    repoNames.push(activeEditor.repoName)
                }
            }

            const unifiedContextFetcherClient = new UnifiedContextFetcherClient(graphqlClient, repoIds)
            const codebaseContext = new CodebaseContext(
                config,
                undefined,
                null,
                null,
                null,
                unifiedContextFetcherClient
            )

            const { humanChatInput = '', prefilledOptions } = options ?? {}
            // TODO(naman): save scope with each interaction
            const interaction = await recipe.getInteraction(humanChatInput, {
                editor: prefilledOptions ? withPreselectedOptions(editor, prefilledOptions) : editor,
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

            const { prompt, contextFiles } = await transcript.getPromptForLastInteraction(
                getMultiRepoPreamble(repoNames)
            )
            transcript.setUsedContextFilesForLastInteraction(contextFiles)

            const responsePrefix = interaction.getAssistantMessage().prefix ?? ''
            let rawText = ''

            const updatedTranscript = await new Promise<Transcript | null>(resolve => {
                const abort = chatClient.chat(prompt, {
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
                        transcript.addErrorAsAssistantResponse(error)

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

                setAbortMessageInProgress(() => () => {
                    abort()
                    resolve(transcript)
                })
            })

            setAbortMessageInProgress(() => () => undefined)

            return updatedTranscript
        },
        [
            config,
            scope,
            codebases,
            codebaseIds,
            graphqlClient,
            transcript,
            intentDetector,
            chatClient,
            isMessageInProgress,
            onEvent,
            setAbortMessageInProgress,
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
            toggleIncludeInferredRepository,
            toggleIncludeInferredFile,
            abortMessageInProgress,
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
            toggleIncludeInferredRepository,
            toggleIncludeInferredFile,
            abortMessageInProgress,
        ]
    )
}
