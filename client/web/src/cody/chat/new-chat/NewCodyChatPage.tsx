import type { FC } from 'react'

import { CodyWebChatProvider, ChatHistory } from 'cody-web-experimental'

import { Badge, Link, PageHeader, Text } from '@sourcegraph/wildcard'

import { Page } from '../../../components/Page'
import { PageTitle } from '../../../components/PageTitle'
import { CodyColorIcon } from '../CodyPageIcon'

import { ChatHistoryList } from './components/chat-history-list/ChatHistoryList'
import { ChatUi } from './components/chat-ui/ChatUi'
import { Skeleton } from './components/skeleton/Skeleton'

import styles from './NewCodyChatPage.module.scss'

interface NewCodyChatPageProps {
    isSourcegraphDotCom: boolean
}

export const NewCodyChatPage: FC<NewCodyChatPageProps> = props => {
    const { isSourcegraphDotCom } = props

    return (
        <Page className={styles.root}>
            <PageTitle title="Cody Web Chat" />

            <CodyPageHeader isSourcegraphDotCom={isSourcegraphDotCom} className={styles.pageHeader} />

            <div className={styles.chatContainer}>
                <CodyWebChatProvider
                    accessToken=""
                    serverEndpoint={window.location.origin}
                    initialContext={{ repositories: [] }}
                >
                    <ChatHistory>
                        {history => (
                            <div className={styles.chatHistory}>
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
                                        onChatSelect={history.selectChat}
                                        onChatDelete={history.deleteChat}
                                        onChatCreate={history.createNewChat}
                                    />
                                )}
                            </div>
                        )}
                    </ChatHistory>
                    <ChatUi className={styles.chat} />
                </CodyWebChatProvider>
            </div>
        </Page>
    )
}

interface CodyPageHeaderProps {
    isSourcegraphDotCom: boolean
    className: string
}

const CodyPageHeader: FC<CodyPageHeaderProps> = props => {
    const { isSourcegraphDotCom, className } = props

    const codyDashboardLink = isSourcegraphDotCom ? '/cody/manage' : '/cody'

    return (
        <PageHeader
            className={className}
            description="Cody answers code questions and writes code for you using your entire codebase and the code graph."
        >
            <PageHeader.Heading as="h2" styleAs="h1">
                <PageHeader.Breadcrumb icon={CodyColorIcon}>
                    <div className="d-inline-flex align-items-center">
                        Cody Chat
                        <Badge variant="info" className="ml-2">
                            Experimental
                        </Badge>
                        <Link to={codyDashboardLink}>
                            <Text className="mb-0 ml-2" size="small">
                                Manage
                            </Text>
                        </Link>
                    </div>
                </PageHeader.Breadcrumb>
            </PageHeader.Heading>
        </PageHeader>
    )
}
