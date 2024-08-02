import { type FC, memo, useCallback, useMemo } from 'react'

import classNames from 'classnames'
import { useLocation } from 'react-router-dom'

import {
    CodyWebChatProvider,
    type InitialContext,
    type CodyWebChatContextClient,
    CodyWebHistory,
} from '@sourcegraph/cody-web'
import { SourcegraphURL } from '@sourcegraph/common'
import { Text, useLocalStorage } from '@sourcegraph/wildcard'

import { getTelemetrySourceClient } from '../../../telemetry'
import { ChatHistoryList } from '../../chat/new-chat/components/chat-history-list/ChatHistoryList'
import { ChatUi } from '../../chat/new-chat/components/chat-ui/ChatUi'
import { Skeleton } from '../../chat/new-chat/components/skeleton/Skeleton'

import styles from './NewCodySidebarWebChat.module.scss'

interface Repository {
    id: string
    name: string
}

interface NewCodySidebarWebChatProps {
    filePath?: string
    repository: Repository
    mode: 'chat' | 'history'
    onChatSelect?: () => void
    onClientCreated?: (client: CodyWebChatContextClient) => void
}

export const NewCodySidebarWebChat: FC<NewCodySidebarWebChatProps> = memo(function CodyWebChat(props) {
    const { filePath, repository, onChatSelect, onClientCreated, mode } = props

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
        <CodyWebChatProvider
            accessToken=""
            chatID={chatID}
            initialContext={contextInfo}
            serverEndpoint={window.location.origin}
            customHeaders={window.context.xhrHeaders}
            telemetryClientName={getTelemetrySourceClient()}
            onNewChatCreated={handleNewChatCreated}
            onClientCreated={onClientCreated}
        >
            <ChatUi className={classNames({ [styles.hidden]: mode !== 'chat' })} />

            <div className={classNames(styles.chatHistory, { [styles.hidden]: mode !== 'history' })}>
                <CodyWebHistory>
                    {history => (
                        <div>
                            {history.loading && (
                                <>
                                    <Skeleton />
                                    <Skeleton />
                                    <Skeleton />
                                </>
                            )}
                            {history.error && <Text>Error: {history.error.message}</Text>}

                            {!history.loading && !history.error && (
                                <ChatHistoryList
                                    chats={history.chats}
                                    isSelectedChat={history.isSelectedChat}
                                    withCreationButton={false}
                                    onChatSelect={chat => {
                                        onChatSelect()
                                        history.selectChat(chat)
                                    }}
                                    onChatDelete={history.deleteChat}
                                    onChatCreate={history.createNewChat}
                                />
                            )}
                        </div>
                    )}
                </CodyWebHistory>
            </div>
        </CodyWebChatProvider>
    )
})
