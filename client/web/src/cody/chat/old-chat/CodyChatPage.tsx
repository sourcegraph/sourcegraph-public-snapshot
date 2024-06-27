import React, { useEffect, useState } from 'react'

import { mdiCogOutline, mdiDelete, mdiDotsVertical, mdiFormatListBulleted, mdiOpenInNew, mdiPlus } from '@mdi/js'
import classNames from 'classnames'
import { Navigate, useLocation, useNavigate } from 'react-router-dom'

import { CodyLogo } from '@sourcegraph/cody-ui'
import { type AuthenticatedUser } from '@sourcegraph/shared/src/auth'
import type { TelemetryV2Props } from '@sourcegraph/shared/src/telemetry'
import {
    Badge,
    Button,
    ButtonLink,
    H4,
    Icon,
    Link,
    Menu,
    MenuButton,
    MenuDivider,
    MenuItem,
    MenuLink,
    MenuList,
    PageHeader,
    Tooltip,
} from '@sourcegraph/wildcard'

import { Page } from '../../../components/Page'
import { PageTitle } from '../../../components/PageTitle'
import { useFeatureFlag } from '../../../featureFlags/useFeatureFlag'
import type { SourcegraphContext } from '../../../jscontext'
import { PageRoutes } from '../../../routes.constants'
import { EventName } from '../../../util/constants'
import { CodyProRoutes } from '../../codyProRoutes'
import { ChatUI } from '../../components/ChatUI'
import { HistoryList } from '../../components/HistoryList'
import { useCodyChat, type CodyChatStore } from '../../useCodyChat'
import { CodyColorIcon } from '../CodyPageIcon'

import styles from './CodyChatPage.module.scss'

interface CodyChatPageProps extends TelemetryV2Props {
    isSourcegraphDotCom: boolean
    authenticatedUser: AuthenticatedUser
    context: Pick<SourcegraphContext, 'externalURL'>
}

const transcriptIdFromUrl = (pathname: string): string | undefined => {
    const serializedID = pathname.split('/').pop()
    if (!serializedID) {
        return
    }

    try {
        return atob(serializedID)
    } catch {
        return
    }
}

const onTranscriptHistoryLoad = (
    loadTranscriptFromHistory: CodyChatStore['loadTranscriptFromHistory'],
    transcriptHistory: CodyChatStore['transcriptHistory'],
    initializeNewChat: CodyChatStore['initializeNewChat']
): void => {
    if (transcriptHistory.length > 0) {
        const transcriptId = transcriptIdFromUrl(window.location.pathname)

        if (transcriptId && transcriptHistory.find(({ id }) => id === transcriptId)) {
            loadTranscriptFromHistory(transcriptId).catch(() => null)
        } else {
            loadTranscriptFromHistory(transcriptHistory[0].id).catch(() => null)
        }
    } else {
        initializeNewChat()
    }
}

export const CodyChatPage: React.FunctionComponent<CodyChatPageProps> = ({
    authenticatedUser,
    isSourcegraphDotCom,
    telemetryRecorder,
}) => {
    const { pathname } = useLocation()
    const navigate = useNavigate()

    // Evaluate a mock feature flag for the purpose of an A/A test. No functionality is affected by this flag.
    const [_codyChatMockTestValue] = useFeatureFlag('cody-chat-mock-test')

    const codyChatStore = useCodyChat({
        userID: authenticatedUser?.id,
        onTranscriptHistoryLoad,
        autoLoadTranscriptFromHistory: false,
        telemetryRecorder,
    })
    const {
        initializeNewChat,
        clearHistory,
        loaded,
        transcript,
        transcriptHistory,
        loadTranscriptFromHistory,
        deleteHistoryItem,
        logTranscriptEvent,
    } = codyChatStore
    useEffect(() => {
        logTranscriptEvent(EventName.CODY_CHAT_PAGE_VIEWED, 'cody.chat', 'view')
    }, [logTranscriptEvent])

    const transcriptId = transcript?.id

    useEffect(() => {
        if (!loaded || !transcriptId || !authenticatedUser || !window.context?.codyEnabledForCurrentUser) {
            return
        }
        const idFromUrl = transcriptIdFromUrl(pathname)

        if (transcriptId !== idFromUrl) {
            navigate(`/cody/chat/${btoa(transcriptId)}`, {
                replace: true,
            })
        }
    }, [transcriptId, loaded, pathname, navigate, authenticatedUser])

    const [showMobileHistory, setShowMobileHistory] = useState<boolean>(false)
    // Close mobile history list when transcript changes
    useEffect(() => {
        setShowMobileHistory(false)
    }, [transcript])

    if (!loaded) {
        return null
    }

    if (!window.context?.codyEnabledForCurrentUser) {
        return <Navigate to={PageRoutes.CodyDashboard} />
    }

    const codyDashboardLink = isSourcegraphDotCom ? CodyProRoutes.Manage : PageRoutes.CodyDashboard

    return (
        <Page className={classNames('d-flex flex-column', styles.page)}>
            <PageTitle title="Cody chat" />
            <PageHeader
                actions={
                    <div className="d-flex flex-gap-1">
                        <ButtonLink variant="link" to={codyDashboardLink}>
                            Editor extensions
                        </ButtonLink>
                        <ButtonLink variant="secondary" to={codyDashboardLink}>
                            Dashboard
                        </ButtonLink>
                        <Button variant="primary" onClick={initializeNewChat}>
                            <Icon aria-hidden={true} svgPath={mdiPlus} />
                            New chat
                        </Button>
                    </div>
                }
                className={styles.pageHeader}
            >
                <PageHeader.Heading as="h2" styleAs="h1">
                    <PageHeader.Breadcrumb icon={CodyColorIcon}>
                        <div className="d-inline-flex align-items-center">Cody Chat</div>
                    </PageHeader.Breadcrumb>
                </PageHeader.Heading>
            </PageHeader>
            {/* Page content */}
            <div className={classNames('row flex-1 overflow-hidden', styles.pageWrapper)}>
                <div className={classNames('col-md-3', styles.sidebarWrapper)}>
                    <div className={styles.sidebarHeader}>
                        <H4>
                            <b>Chats</b>
                        </H4>
                        <Menu>
                            <MenuButton variant="icon" outline={false}>
                                <Icon aria-hidden={true} svgPath={mdiDotsVertical} size="md" />
                            </MenuButton>

                            <MenuList>
                                {(transcriptHistory.length > 1 || !!transcriptHistory[0]?.interactions?.length) && (
                                    <>
                                        <MenuItem onSelect={clearHistory}>
                                            <Icon aria-hidden={true} svgPath={mdiDelete} /> Clear all chats
                                        </MenuItem>
                                        <MenuDivider />
                                    </>
                                )}
                                <MenuLink as={Link} to="/help/cody" target="_blank" rel="noopener">
                                    <Icon aria-hidden={true} svgPath={mdiOpenInNew} /> Cody Docs & FAQ
                                </MenuLink>
                                {authenticatedUser?.siteAdmin && (
                                    <MenuLink as={Link} to="/site-admin/cody">
                                        <Icon aria-hidden={true} svgPath={mdiCogOutline} /> Cody Settings
                                    </MenuLink>
                                )}
                            </MenuList>
                        </Menu>
                    </div>
                    <div className={classNames('h-100 mb-4', styles.sidebar)}>
                        <HistoryList
                            currentTranscript={transcript}
                            transcriptHistory={transcriptHistory}
                            truncateMessageLength={60}
                            loadTranscriptFromHistory={loadTranscriptFromHistory}
                            deleteHistoryItem={deleteHistoryItem}
                        />
                    </div>
                </div>

                <div className={classNames('col-md-9 h-100', styles.chatMainWrapper)}>
                    <div className={styles.mobileTopBarWeb}>
                        <div className="d-flex col-2 p-0">
                            <Tooltip content="Chat history">
                                <Button
                                    variant="icon"
                                    className="mr-2"
                                    aria-label="Active chats"
                                    onClick={() => setShowMobileHistory(true)}
                                    aria-pressed={showMobileHistory}
                                >
                                    <Icon aria-hidden={true} svgPath={mdiFormatListBulleted} />
                                </Button>
                            </Tooltip>
                            <Tooltip content="Start a new chat">
                                <Button variant="icon" aria-label="Start a new chat" onClick={initializeNewChat}>
                                    <Icon aria-hidden={true} svgPath={mdiPlus} />
                                </Button>
                            </Tooltip>
                            {(transcriptHistory.length > 1 || !!transcriptHistory[0]?.interactions?.length) && (
                                <Tooltip content="Clear all chats">
                                    <Button
                                        variant="icon"
                                        className="ml-2"
                                        aria-label="Clear all chats"
                                        onClick={clearHistory}
                                    >
                                        <Icon aria-hidden={true} svgPath={mdiDelete} />
                                    </Button>
                                </Tooltip>
                            )}
                        </div>
                        <div className="col-8 d-flex justify-content-center">
                            <div className="d-flex flex-shrink-0 align-items-center">
                                <CodyLogo />
                                {showMobileHistory ? 'Chats' : 'Ask Cody'}
                                <div className="ml-2">
                                    <Badge variant="info">Experimental</Badge>
                                </div>
                            </div>
                        </div>
                        <div className="col-2 d-flex" />
                    </div>
                    {showMobileHistory ? (
                        <HistoryList
                            currentTranscript={transcript}
                            transcriptHistory={transcriptHistory}
                            truncateMessageLength={60}
                            loadTranscriptFromHistory={loadTranscriptFromHistory}
                            deleteHistoryItem={deleteHistoryItem}
                        />
                    ) : (
                        <ChatUI
                            codyChatStore={codyChatStore}
                            isCodyChatPage={true}
                            authenticatedUser={authenticatedUser}
                            telemetryRecorder={telemetryRecorder}
                        />
                    )}
                </div>
            </div>
        </Page>
    )
}
