import create from 'zustand'

import { Client, createClient, ClientInit } from '@sourcegraph/cody-shared/src/chat/client'
import { ChatMessage } from '@sourcegraph/cody-shared/src/chat/transcript/messages'
import { isErrorLike } from '@sourcegraph/common'

import { eventLogger } from '../tracking/eventLogger'

interface CodyChatStore {
    client: Client | null
    messageInProgress: ChatMessage | null
    transcript: ChatMessage[]
    filePath: string
    setClient: (client: Client | null) => void
    setMessageInProgress: (message: ChatMessage | null) => void
    setTranscript: (transcript: ChatMessage[]) => void
    initializeClient: (config: ClientInit['config']) => void
    onSubmit: (text: string) => void
}

export const useChatStoreState = create<CodyChatStore>((set, get) => {
    const onSubmit = (text: string) => {
        const client = get().client
        if (client && !isErrorLike(client)) {
            // eventLogger.log('web:codySidebar:submit', { repo: get().config?.codebase, path: get().filePath, text })
            client.submitMessage(text)
        }
    }

    return {
        client: null,
        messageInProgress: null,
        transcript: [],
        filePath: '',
        setClient: client => set({ client }),
        setMessageInProgress: message => set({ messageInProgress: message }),
        setTranscript: transcript => set({ transcript: transcript }),

        initializeClient: config => {
            set({ messageInProgress: null })
            set({ transcript: [] })
            createClient({
                config: config,
                accessToken: null,
                setMessageInProgress: message => set({ messageInProgress: message }),
                setTranscript: transcript => set({ transcript: transcript }),
            })
                .then(client => {
                    set({ client })
                })
                .catch(error => {
                    eventLogger.log('web:codySidebar:clientError', { repo: config?.codebase })
                    set({ client: error })
                })
        },

        onSubmit,
    }
})
