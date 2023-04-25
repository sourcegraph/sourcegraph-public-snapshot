/* eslint-disable no-void */
import { useCallback, useEffect, useMemo, useRef } from 'react'

import { isEqual } from 'lodash'
import create from 'zustand'

import { Client, createClient, ClientInit, Transcript, TranscriptJSON } from '@sourcegraph/cody-shared/src/chat/client'
import { ChatContextStatus } from '@sourcegraph/cody-shared/src/chat/context'
import { ChatMessage } from '@sourcegraph/cody-shared/src/chat/transcript/messages'
import { PrefilledOptions } from '@sourcegraph/cody-shared/src/editor/withPreselectedOptions'
import { isErrorLike } from '@sourcegraph/common'

import { CodeMirrorEditor } from '../cody/CodeMirrorEditor'
import { useFeatureFlag } from '../featureFlags/useFeatureFlag'
import { eventLogger } from '../tracking/eventLogger'

import { EditorStore, useEditorStore } from './editor'

interface CodyChatStore {
    readonly client: Client | null
    readonly config: ClientInit['config'] | null
    readonly editor: CodeMirrorEditor | null
    readonly messageInProgress: ChatMessage | null
    readonly transcript: ChatMessage[]
    readonly transcriptHistory: TranscriptJSON[]
    // private, not used outside of this module
    onEvent: ((eventName: 'submit' | 'reset' | 'error') => void) | null
    initializeClient: (
        config: Required<ClientInit['config']>,
        editorStore: React.MutableRefObject<EditorStore>,
        onEvent: (eventName: 'submit' | 'reset' | 'error') => void
    ) => Promise<void>
    submitMessage: (text: string) => void
    editMessage: (text: string) => void
    executeRecipe: (
        recipeId: string,
        options?: {
            prefilledOptions?: PrefilledOptions
        }
    ) => Promise<void>
    reset: () => void
    getChatContext: () => ChatContextStatus
}

const CODY_TRANSCRIPT_HISTORY_KEY = 'cody:transcript-history'

// TODO(naman):
// 1. make transcriptHistory database backed
// 2. integrate editor and full filePath context from it

export const useChatStoreState = create<CodyChatStore>((set, get): CodyChatStore => {
    const submitMessage = (text: string): void => {
        const { client, onEvent, getChatContext } = get()
        if (client && !isErrorLike(client)) {
            const { codebase, filePath } = getChatContext()
            eventLogger.log('web:codySidebar:submit', {
                repo: codebase,
                path: filePath,
                text,
            })
            onEvent?.('submit')
            void client.submitMessage(text)
        }
    }

    const editMessage = (text: string): void => {
        const { client, onEvent, getChatContext } = get()
        if (client && !isErrorLike(client)) {
            const { codebase, filePath } = getChatContext()
            eventLogger.log('web:codySidebar:edit', {
                repo: codebase,
                path: filePath,
                text,
            })
            onEvent?.('submit')
            client.transcript.removeLastInteraction()
            void client.submitMessage(text)
        }
    }

    const executeRecipe = async (
        recipeId: string,
        options?: {
            prefilledOptions?: PrefilledOptions
        }
    ): Promise<void> => {
        const { client, getChatContext, onEvent } = get()
        if (client && !isErrorLike(client)) {
            const { codebase, filePath } = getChatContext()
            eventLogger.log('web:codySidebar:recipe', { repo: codebase, path: filePath, recipeId })
            onEvent?.('submit')
            await client.executeRecipe(recipeId, options)
            eventLogger.log('web:codySidebar:recipe:executed', { repo: codebase, path: filePath, recipeId })
        }
        return Promise.resolve()
    }

    const reset = async (): Promise<void> => {
        const { client, onEvent, transcriptHistory } = get()

        if (client && !isErrorLike(client)) {
            // push current transcript to transcript history and save to localstorage
            const transcript = await client.transcript.toJSON()
            if (transcript.interactions.length) {
                transcriptHistory.push(transcript)
            }
            window.localStorage.setItem(CODY_TRANSCRIPT_HISTORY_KEY, JSON.stringify(transcriptHistory))
            set({ transcriptHistory: [...transcriptHistory], messageInProgress: null, transcript: [] })

            onEvent?.('reset')
            void client.reset()
        }
    }

    const setTranscript = async (transcript: ChatMessage[]): Promise<void> => {
        const { transcriptHistory, client } = get()

        if (client && !isErrorLike(client)) {
            set({ transcript })

            const transcriptJSON = await client.transcript.toJSON()

            if (!transcriptHistory.length) {
                transcriptHistory.push(transcriptJSON)
            } else {
                transcriptHistory[transcriptHistory.length - 1] = transcriptJSON
            }

            window.localStorage.setItem(CODY_TRANSCRIPT_HISTORY_KEY, JSON.stringify(transcriptHistory))
            set({ transcriptHistory: [...transcriptHistory] })
        }
    }

    const setMessageInProgress = (message: ChatMessage | null): void => set({ messageInProgress: message })

    const initializeClient = async (
        config: Required<ClientInit['config']>,
        editorStateRef: React.MutableRefObject<EditorStore>,
        onEvent: (eventName: 'submit' | 'reset' | 'error') => void
    ): Promise<void> => {
        const editor = new CodeMirrorEditor(editorStateRef)

        const transcriptHistory = ((): TranscriptJSON[] => {
            try {
                return JSON.parse(window.localStorage.getItem(CODY_TRANSCRIPT_HISTORY_KEY) || '[]')
            } catch {
                return []
            }
        })()

        const initialTranscript = ((): Transcript => {
            try {
                return Transcript.fromJSON(transcriptHistory[transcriptHistory.length - 1] || { interactions: [] })
            } catch {
                return new Transcript()
            }
        })()

        set({
            config,
            editor,
            onEvent,
            transcript: initialTranscript.toChat(),
            transcriptHistory,
        })

        try {
            const client = await createClient({
                config,
                editor,
                setMessageInProgress,
                initialTranscript,
                setTranscript: (transcript: ChatMessage[]): void => void setTranscript(transcript),
            })

            set({ client })
        } catch (error) {
            eventLogger.log('web:codySidebar:clientError', { repo: config?.codebase })
            onEvent('error')
            set({ client: error })
        }
    }

    const getChatContext = (): ChatContextStatus => {
        const { config, editor } = get()

        return {
            codebase: config?.codebase,
            filePath: editor?.getActiveTextEditorSelectionOrEntireFile()?.fileName,
            supportsKeyword: false,
            mode: config?.useContext,
            connection: true,
        }
    }

    return {
        client: null,
        editor: null,
        messageInProgress: null,
        config: null,
        transcript: [],
        transcriptHistory: [],
        onEvent: null,
        initializeClient,
        submitMessage,
        editMessage,
        executeRecipe,
        reset: () => void reset(),
        getChatContext,
    }
})

export const useChatStore = ({
    codebase,
    setIsCodySidebarOpen,
}: {
    codebase: string
    setIsCodySidebarOpen: (state: boolean | undefined) => void
}): CodyChatStore => {
    const [isCodyEnabled] = useFeatureFlag('cody-experimental')
    const store = useChatStoreState()

    const onEvent = useCallback(
        (eventName: 'submit' | 'reset' | 'error') => {
            if (eventName === 'submit') {
                setIsCodySidebarOpen(true)
            }
        },
        [setIsCodySidebarOpen]
    )

    // We use a ref here so that a change in the editor state does not need a recreation of the
    // client config.
    const editorStore = useEditorStore()
    const editorStateRef = useRef(editorStore)
    useEffect(() => {
        editorStateRef.current = editorStore
    }, [editorStore])

    // TODO(naman): change useContext to `blended` after adding keyboard context
    const config = useMemo<Required<ClientInit['config']>>(
        () => ({
            serverEndpoint: window.location.origin,
            useContext: 'embeddings',
            codebase,
            accessToken: null,
        }),
        [codebase]
    )

    const { initializeClient, config: currentConfig } = store
    useEffect(() => {
        if (!isCodyEnabled || isEqual(config, currentConfig)) {
            return
        }

        void initializeClient(config, editorStateRef, onEvent)
    }, [config, initializeClient, currentConfig, isCodyEnabled, editorStateRef, onEvent])

    return store
}
