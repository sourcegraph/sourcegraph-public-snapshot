import { type FC, memo, useCallback, useMemo } from 'react'

import { useLocation } from 'react-router-dom'

import { CodyWebPanelProvider, type InitialContext } from '@sourcegraph/cody-web'
import { SourcegraphURL } from '@sourcegraph/common'
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

    const location = useLocation()
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

    const contextInfo = useMemo<InitialContext>(() => {
        const lineOrPosition = SourcegraphURL.from(location).lineRange
        const hasFileRangeSelection = lineOrPosition.line

        return {
            repositories: [repository],
            fileURL: filePath ? `/${filePath}` : undefined,
            // Line range - 1 because of Cody Web initial context file range bug
            fileRange: hasFileRangeSelection
                ? {
                      startLine: lineOrPosition.line - 1,
                      endLine: lineOrPosition.endLine ? lineOrPosition.endLine - 1 : lineOrPosition.line - 1,
                  }
                : undefined,
        }
    }, [repository, filePath, location])

    const chatID = contextToChatIds[`${repository.id}-${filePath}`] ?? null

    return (
        <CodyWebPanelProvider
            accessToken=""
            chatID={chatID}
            initialContext={contextInfo}
            serverEndpoint={window.location.origin}
            customHeaders={window.context.xhrHeaders}
            telemetryClientName={getTelemetrySourceClient()}
            onNewChatCreated={handleNewChatCreated}
        >
            <ChatUi />
        </CodyWebPanelProvider>
    )
})
