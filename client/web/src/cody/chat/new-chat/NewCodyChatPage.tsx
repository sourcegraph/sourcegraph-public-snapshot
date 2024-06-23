import type { FC } from 'react'

import { ChatHistory, CodyWebChatProvider } from 'cody-web-experimental'
import { Navigate } from 'react-router-dom'

import { Badge, ButtonLink, PageHeader, Text } from '@sourcegraph/wildcard'

import { Page } from '../../../components/Page'
import { PageTitle } from '../../../components/PageTitle'
import { PageRoutes } from '../../../routes.constants'
import { CodyProRoutes } from '../../codyProRoutes'
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

    const codyDashboardLink = isSourcegraphDotCom ? CodyProRoutes.Manage : PageRoutes.CodyDashboard

    if (!window.context?.codyEnabledForCurrentUser) {
        return <Navigate to={PageRoutes.CodyDashboard} />
    }

    return (
        <PageHeader
            className={className}
            actions={
                <div className="d-flex flex-gap-1">
                    <ButtonLink variant="link" to={codyDashboardLink}>
                        Editor extensions
                    </ButtonLink>
                    <ButtonLink variant="secondary" to={codyDashboardLink}>
                        Dashboard
                    </ButtonLink>
                </div>
            }
        >
            <PageHeader.Heading as="h2" styleAs="h1">
                <PageHeader.Breadcrumb icon={CodyColorIcon}>
                    <div className="d-inline-flex align-items-center">
                        Cody Chat
                        <Badge variant="info" className="ml-2">
                            Experimental
                        </Badge>
                    </div>
                </PageHeader.Breadcrumb>
            </PageHeader.Heading>
        </PageHeader>
    )
}
