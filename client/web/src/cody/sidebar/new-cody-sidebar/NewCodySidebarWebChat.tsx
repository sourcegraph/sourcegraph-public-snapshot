import { type FC, memo, useCallback, useMemo } from 'react'

import { CodyWebChatProvider } from '@sourcegraph/cody-web'
import { useLocalStorage } from '@sourcegraph/wildcard'

import { getTelemetrySourceClient } from '../../../telemetry'
import { ChatUi } from '../../chat/new-chat/components/chat-ui/ChatUi'

interface Repository {
    id: string
    name: string
}

interface NewCodySidebarWebChatProps {
    filePath?: string
    repository: Repository
}

export const NewCodySidebarWebChat: FC<NewCodySidebarWebChatProps> = memo(function CodyWebChat(props) {
    const { filePath, repository } = props

    const [contextToChatIds, setContextToChatIds] = useLocalStorage<Record<string, string>>(
        'cody.context-to-chat-ids',
        {}
    )

    const handleNewChatCreated = useCallback(
        (chatId: string): void => {
            contextToChatIds[`${repository.id}-${filePath}`] = chatId
            setContextToChatIds(contextToChatIds)
        },
        [contextToChatIds, setContextToChatIds, filePath, repository.id]
    )

    const contextInfo = useMemo(
        () => ({
            repositories: [repository],
            fileURL: filePath ? `/${filePath}` : undefined,
        }),
        [repository, filePath]
    )

    const chatID = contextToChatIds[`${repository.id}-${filePath}`] ?? null

    return (
        <CodyWebChatProvider
            accessToken=""
            chatID={chatID}
            initialContext={contextInfo}
            serverEndpoint={window.location.origin}
            customHeaders={window.context.xhrHeaders}
            telemetryClientName={getTelemetrySourceClient()}
            onNewChatCreated={handleNewChatCreated}
        >
            <ChatUi />
        </CodyWebChatProvider>
    )
})
