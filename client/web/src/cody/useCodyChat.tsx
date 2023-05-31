import { useState, useEffect, useCallback, useMemo } from 'react'

import { Transcript, TranscriptJSON } from '@sourcegraph/cody-shared/src/chat/transcript'
import {
    useClient,
    CodyClient,
    CodyClientScope,
    CodyClientConfig,
    CodyClientEvent,
} from '@sourcegraph/cody-shared/src/chat/useClient'
import { useLocalStorage } from '@sourcegraph/wildcard'

import { CodeMirrorEditor } from './components/CodeMirrorEditor'
import { useIsCodyEnabled, IsCodyEnabled, notEnabled } from './useIsCodyEnabled'

export type { CodyClientScope } from '@sourcegraph/cody-shared/src/chat/useClient'

export interface CodyChatStore
    extends Pick<
        CodyClient,
        | 'transcript'
        | 'chatMessages'
        | 'isMessageInProgress'
        | 'messageInProgress'
        | 'submitMessage'
        | 'editMessage'
        | 'initializeNewChat'
        | 'executeRecipe'
        | 'scope'
        | 'setScope'
        | 'setEditorScope'
        | 'legacyChatContext'
    > {
    readonly transcriptHistory: TranscriptJSON[]
    readonly loaded: boolean
    readonly isCodyEnabled: IsCodyEnabled
    clearHistory: () => void
    deleteHistoryItem: (id: string) => void
    loadTranscriptFromHistory: (id: string) => Promise<void>
}

export const codyChatStoreMock: CodyChatStore = {
    transcript: null,
    chatMessages: [],
    isMessageInProgress: false,
    messageInProgress: null,
    submitMessage: () => Promise.resolve(null),
    editMessage: () => Promise.resolve(null),
    initializeNewChat: () => null,
    executeRecipe: () => Promise.resolve(null),
    scope: { type: 'Automatic', repositories: [], editor: new CodeMirrorEditor() },
    setScope: () => {},
    setEditorScope: () => {},
    legacyChatContext: {},
    transcriptHistory: [],
    loaded: true,
    isCodyEnabled: notEnabled,
    clearHistory: () => {},
    deleteHistoryItem: () => {},
    loadTranscriptFromHistory: () => Promise.resolve(),
}

interface CodyChatProps {
    scope?: CodyClientScope
    config?: CodyClientConfig
    onEvent?: (event: CodyClientEvent) => void
    onTranscriptHistoryLoad?: (
        loadTranscriptFromHistory: (id: string) => Promise<void>,
        transcriptHistory: TranscriptJSON[],
        initializeNewChat: CodyClient['initializeNewChat']
    ) => void
    autoLoadTranscriptFromHistory?: boolean
}

const CODY_TRANSCRIPT_HISTORY_KEY = 'cody.chat.history'
const SAVE_MAX_TRANSCRIPT_HISTORY = 20

export const useCodyChat = ({
    scope: initialScope,
    config: initialConfig,
    onEvent,
    onTranscriptHistoryLoad,
    autoLoadTranscriptFromHistory = true,
}: CodyChatProps): CodyChatStore => {
    const isCodyEnabled = useIsCodyEnabled()
    const [loadedTranscriptFromHistory, setLoadedTranscriptFromHistory] = useState(false)
    const [transcriptHistoryInternal, setTranscriptHistoryState] = useLocalStorage<TranscriptJSON[]>(
        CODY_TRANSCRIPT_HISTORY_KEY,
        []
    )
    const transcriptHistory = useMemo(() => transcriptHistoryInternal || [], [transcriptHistoryInternal])

    const {
        transcript,
        isMessageInProgress,
        messageInProgress,
        chatMessages,
        scope,
        setScope,
        setEditorScope,
        setTranscript,
        legacyChatContext,
        initializeNewChat: initializeNewChatInternal,
        submitMessage: submitMessageInternal,
        editMessage: editMessageInternal,
        executeRecipe: executeRecipeInternal,
        ...client
    } = useClient({
        config: initialConfig || {
            serverEndpoint: window.location.origin,
            useContext: 'embeddings',
            accessToken: null,
            customHeaders: window.context.xhrHeaders,
            debugEnable: false,
            needsEmailVerification: isCodyEnabled.needsEmailVerification,
        },
        scope: initialScope,
        onEvent,
        web: true,
    })

    const loadTranscriptFromHistory = useCallback(
        async (id: string) => {
            if (transcript?.id === id) {
                return
            }

            const transcriptToLoad = transcriptHistory.find(item => item.id === id)
            if (transcriptToLoad) {
                await setTranscript(Transcript.fromJSON(transcriptToLoad))
            }
        },
        [transcriptHistory, transcript?.id, setTranscript]
    )

    const clearHistory = useCallback(() => {
        if (client.config.needsEmailVerification) {
            return
        }

        const newTranscript = initializeNewChatInternal()
        if (newTranscript) {
            setTranscriptHistoryState([newTranscript.toJSONEmpty()])
        } else {
            setTranscriptHistoryState([])
        }
    }, [client.config.needsEmailVerification, initializeNewChatInternal, setTranscriptHistoryState])

    const deleteHistoryItem = useCallback(
        (id: string): void => {
            if (client.config.needsEmailVerification) {
                return
            }

            setTranscriptHistoryState((history: TranscriptJSON[]) => {
                const updatedHistory = [...history.filter(transcript => transcript.id !== id)]

                if (transcript?.id === id) {
                    if (updatedHistory.length === 0) {
                        const newTranscript = initializeNewChatInternal()

                        if (newTranscript) {
                            updatedHistory.push(newTranscript.toJSONEmpty())
                        }
                    } else {
                        setTranscript(Transcript.fromJSON(updatedHistory[0])).catch(() => null)
                    }
                }

                return sortSliceTranscriptHistory(updatedHistory)
            })
        },
        [
            setTranscript,
            client.config.needsEmailVerification,
            initializeNewChatInternal,
            transcript?.id,
            setTranscriptHistoryState,
        ]
    )

    const updateTranscriptInHistory = useCallback(
        async (transcript: Transcript) => {
            const transcriptJSON = await transcript.toJSON()

            setTranscriptHistoryState((history: TranscriptJSON[]) => {
                const index = history.findIndex(item => item.id === transcript.id)
                if (index >= 0) {
                    history[index] = transcriptJSON
                }

                return [...history]
            })
        },
        [setTranscriptHistoryState]
    )

    const pushTranscriptToHistory = useCallback(
        async (transcript: Transcript) => {
            const transcriptJSON = await transcript.toJSON()

            setTranscriptHistoryState((history: TranscriptJSON[] = []) =>
                sortSliceTranscriptHistory([transcriptJSON, ...history])
            )
        },
        [setTranscriptHistoryState]
    )

    const submitMessage = useCallback<typeof submitMessageInternal>(
        async (humanInputText, scope): Promise<Transcript | null> => {
            const transcript = await submitMessageInternal(humanInputText, scope)

            if (transcript) {
                await updateTranscriptInHistory(transcript)
            }

            return transcript
        },
        [submitMessageInternal, updateTranscriptInHistory]
    )
    const editMessage = useCallback<typeof editMessageInternal>(
        async (humanInputText, messageId?, scope?): Promise<Transcript | null> => {
            const transcript = await editMessageInternal(humanInputText, messageId, scope)

            if (transcript) {
                await updateTranscriptInHistory(transcript)
            }

            return transcript
        },
        [editMessageInternal, updateTranscriptInHistory]
    )

    const initializeNewChat = useCallback((): Transcript | null => {
        const transcript = initializeNewChatInternal()

        if (transcript) {
            pushTranscriptToHistory(transcript).catch(() => null)
        }

        return transcript
    }, [initializeNewChatInternal, pushTranscriptToHistory])

    const executeRecipe = useCallback<typeof executeRecipeInternal>(
        async (recipeId, options): Promise<Transcript | null> => {
            const transcript = await executeRecipeInternal(recipeId, options)

            if (transcript) {
                await updateTranscriptInHistory(transcript)
            }

            return transcript
        },
        [executeRecipeInternal, updateTranscriptInHistory]
    )

    const loaded = useMemo(
        () => loadedTranscriptFromHistory && isCodyEnabled.loaded,
        [loadedTranscriptFromHistory, isCodyEnabled.loaded]
    )

    // Autoload the latest transcript from history once it is loaded. Initially the transcript is null.
    useEffect(() => {
        if (!loadedTranscriptFromHistory && transcript === null) {
            const history = sortSliceTranscriptHistory([...transcriptHistory])

            if (autoLoadTranscriptFromHistory) {
                if (history.length > 0) {
                    setTranscript(Transcript.fromJSON(history[0])).catch(() => null)
                } else {
                    const newTranscript = new Transcript()
                    history.push({ interactions: [], id: newTranscript.id, lastInteractionTimestamp: newTranscript.id })
                    setTranscript(newTranscript)
                        .then(() => setTranscriptHistoryState(history))
                        .catch(() => null)
                }
            }
            // usefull to load transcript from any other source like url.
            onTranscriptHistoryLoad?.(loadTranscriptFromHistory, history, initializeNewChat)

            setLoadedTranscriptFromHistory(true)
        }
    }, [
        transcriptHistory,
        loadedTranscriptFromHistory,
        transcript,
        autoLoadTranscriptFromHistory,
        onTranscriptHistoryLoad,
        setTranscript,
        setTranscriptHistoryState,
        loadTranscriptFromHistory,
        initializeNewChat,
    ])

    return {
        loaded,
        isCodyEnabled,
        transcript,
        transcriptHistory,
        chatMessages,
        messageInProgress,
        isMessageInProgress,
        submitMessage,
        editMessage,
        initializeNewChat,
        clearHistory,
        deleteHistoryItem,
        executeRecipe,
        scope,
        setScope,
        setEditorScope,
        loadTranscriptFromHistory,
        legacyChatContext,
    }
}

export const safeTimestampToDate = (timestamp: string = ''): Date => {
    if (isNaN(new Date(timestamp) as any)) {
        return new Date()
    }

    return new Date(timestamp)
}

const sortSliceTranscriptHistory = (transcriptHistory: TranscriptJSON[]): TranscriptJSON[] =>
    transcriptHistory
        .sort(
            (a, b) =>
                (safeTimestampToDate(b.lastInteractionTimestamp) as any) -
                (safeTimestampToDate(a.lastInteractionTimestamp) as any)
        )
        .map(transcript => (transcript.id ? transcript : { ...transcript, id: Transcript.fromJSON(transcript).id }))
        .slice(0, SAVE_MAX_TRANSCRIPT_HISTORY)
