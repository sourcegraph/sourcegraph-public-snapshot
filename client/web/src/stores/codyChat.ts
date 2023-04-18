/* eslint-disable no-void */
import { useEffect, useMemo, useRef } from 'react'

import create from 'zustand'

import { Client, createClient, ClientInit } from '@sourcegraph/cody-shared/src/chat/client'
import { ChatMessage } from '@sourcegraph/cody-shared/src/chat/transcript/messages'
import { PrefilledOptions } from '@sourcegraph/cody-shared/src/editor/withPreselectedOptions'
import { isErrorLike } from '@sourcegraph/common'

import { CodeMirrorEditor } from '../cody/CodeMirrorEditor'
import { eventLogger } from '../tracking/eventLogger'

import { EditorStore, useEditorStore } from './editor'

interface CodyChatStore {
    client: Client | null
    config: ClientInit['config'] | null
    messageInProgress: ChatMessage | null
    transcript: ChatMessage[]
    repo: string
    filePath: string

    initializeClient: (
        config: Required<ClientInit['config']>,
        editorStore: React.MutableRefObject<EditorStore>,
        openCody: () => void
    ) => Promise<void>

    submitMessage: (text: string) => void
    executeRecipe: (
        recipeId: string,
        options?: {
            prefilledOptions?: PrefilledOptions
        }
    ) => Promise<void>
    reset: () => void
}

export const useChatStoreState = create<CodyChatStore>((set, get): CodyChatStore => {
    const submitMessage = (text: string): void => {
        const { client, repo, filePath } = get()
        if (client && !isErrorLike(client)) {
            eventLogger.log('web:codySidebar:submit', {
                repo,
                path: filePath,
                text,
            })
            void client.submitMessage(text)
        }
    }

    const executeRecipe = async (
        recipeId: string,
        options?: {
            prefilledOptions?: PrefilledOptions
        }
    ): Promise<void> => {
        const { client, repo, filePath } = get()
        if (client && !isErrorLike(client)) {
            eventLogger.log('web:codySidebar:recipe', { repo, path: filePath, recipeId })
            await client.executeRecipe(recipeId, options)
            eventLogger.log('web:codySidebar:recipe:executed', { repo, path: filePath, recipeId })
        }
        return Promise.resolve()
    }

    const reset = (): void => {
        const { client } = get()
        if (client && !isErrorLike(client)) {
            void client.reset()
        }
    }

    return {
        client: null,
        messageInProgress: null,
        config: null,
        transcript: [],
        filePath: '',
        repo: '',

        async initializeClient(
            config: Required<ClientInit['config']>,
            stateRef: React.MutableRefObject<EditorStore>,
            openCody: () => void
        ): Promise<void> {
            set({ messageInProgress: null, transcript: [], repo: config.codebase, config })

            const editor = new CodeMirrorEditor(stateRef)

            try {
                const client = await createClient({
                    config,
                    setMessageInProgress: message => set({ messageInProgress: message }),
                    setTranscript: transcript => set({ transcript }),
                    editor,
                    openCody,
                })

                set({ client })
            } catch (error) {
                set({ client: error })
            }
        },

        submitMessage,
        executeRecipe,
        reset,
    }
})

export const useChatStore = (isCodyEnabled: boolean, repoName: string, openCody: () => void): CodyChatStore => {
    const store = useChatStoreState()

    const editorStore = useEditorStore()
    // We use a ref here so that a change in the editor state does not need a recreation of the
    // client config.
    const stateRef = useRef(editorStore)
    useEffect(() => {
        stateRef.current = editorStore
    }, [editorStore])

    const config = useMemo<Required<ClientInit['config']>>(
        () => ({
            serverEndpoint: window.location.origin,
            useContext: 'embeddings',
            codebase: repoName,
            accessToken: null,
        }),
        [repoName]
    )

    const { initializeClient } = store
    useEffect(() => {
        if (!isCodyEnabled) {
            return
        }

        void initializeClient(config, stateRef, openCody)
    }, [config, initializeClient, isCodyEnabled, stateRef, openCody])

    return store
}
