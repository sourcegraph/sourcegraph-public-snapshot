import React, { useEffect, useState } from 'react'

import {
    mdiClose,
    mdiCogOutline,
    mdiDelete,
    mdiDotsVertical,
    mdiOpenInNew,
    mdiPlus,
    mdiChevronRight,
    mdiViewList,
    mdiFormatListBulleted,
} from '@mdi/js'
import classNames from 'classnames'
import { useLocation, useNavigate } from 'react-router-dom'

import { CodyLogo } from '@sourcegraph/cody-ui/src/icons/CodyLogo'
import { AuthenticatedUser } from '@sourcegraph/shared/src/auth'
import { useTemporarySetting } from '@sourcegraph/shared/src/settings/temporary'
import {
    Alert,
    Badge,
    Button,
    Icon,
    Menu,
    MenuButton,
    MenuList,
    MenuDivider,
    MenuItem,
    MenuLink,
    PageHeader,
    Link,
    H4,
    H3,
    Text,
    Tooltip,
} from '@sourcegraph/wildcard'

import { MarketingBlock } from '../../components/MarketingBlock'
import { Page } from '../../components/Page'
import { PageTitle } from '../../components/PageTitle'
import { SourcegraphContext } from '../../jscontext'
import { eventLogger } from '../../tracking/eventLogger'
import { EventName } from '../../util/constants'
import { ChatUI } from '../components/ChatUI'
import { CodyMarketingPage } from '../components/CodyMarketingPage'
import { HistoryList } from '../components/HistoryList'
import { CodyChatStore, useCodyChat } from '../useCodyChat'

import { CodyColorIcon } from './CodyPageIcon'

import styles from './CodyChatPage.module.scss'

interface CodyChatPageProps {
    authenticatedUser: AuthenticatedUser | null
    isSourcegraphApp: boolean
    context: Pick<SourcegraphContext, 'authProviders'>
}

const onDownloadVSCodeClick = (): void => eventLogger.log(EventName.CODY_CHAT_DOWNLOAD_VSCODE)
const onTryOnPublicCodeClick = (): void => eventLogger.log(EventName.CODY_CHAT_TRY_ON_PUBLIC_CODE)

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
    context,
    isSourcegraphApp,
}) => {
    const { pathname } = useLocation()
    const navigate = useNavigate()

    const codyChatStore = useCodyChat({
        onTranscriptHistoryLoad,
        autoLoadTranscriptFromHistory: false,
    })
    const {
        initializeNewChat,
        clearHistory,
        isCodyEnabled,
        loaded,
        transcript,
        transcriptHistory,
        loadTranscriptFromHistory,
        deleteHistoryItem,
    } = codyChatStore
    const [showVSCodeCTA] = useState<boolean>(Math.random() < 0.5 || true)
    const [isCTADismissed = true, setIsCTADismissed] = useTemporarySetting('cody.chatPageCta.dismissed', false)
    const onCTADismiss = (): void => setIsCTADismissed(true)

    useEffect(() => {
        eventLogger.log(EventName.CODY_CHAT_PAGE_VIEWED)
    }, [])

    const transcriptId = transcript?.id

    useEffect(() => {
        if (!loaded || !transcriptId) {
            return
        }
        const idFromUrl = transcriptIdFromUrl(pathname)

        if (transcriptId !== idFromUrl) {
            navigate(`/cody/${btoa(transcriptId)}`, {
                replace: true,
            })
        }
    }, [transcriptId, loaded, pathname, navigate])

    const [showMobileHistory, setShowMobileHistory] = useState<boolean>(false)
    // Close mobile history list when transcript changes
    useEffect(() => {
        setShowMobileHistory(false)
    }, [transcript])

    if (!loaded) {
        return null
    }

    if (!authenticatedUser) {
        return <CodyMarketingPage context={context} />
    }

    if (!isCodyEnabled.chat) {
        return (
            <Page className="overflow-hidden">
                <PageTitle title="Cody AI Chat" />
                <Alert variant="info">Cody is not enabled. Please contact your site admin to enable Cody.</Alert>
            </Page>
        )
    }

    return (
        <Page className={classNames('d-flex flex-column', styles.page)}>
            <PageTitle title="Cody AI Chat" />
            <PageHeader
                actions={
                    <div className="d-flex">
                        <Button variant="primary" onClick={initializeNewChat}>
                            <Icon aria-hidden={true} svgPath={mdiPlus} />
                            New chat
                        </Button>
                    </div>
                }
                description={
                    <>
                        Cody answers code questions and writes code for you by reading your entire codebase and the code
                        graph.
                    </>
                }
                className={styles.pageHeader}
            >
                <PageHeader.Heading as="h2" styleAs="h1">
                    <PageHeader.Breadcrumb icon={CodyColorIcon}>
                        <div className="d-inline-flex align-items-center">
                            Cody Chat
                            {!isSourcegraphApp && (
                                <Badge variant="info" className="ml-2">
                                    Beta
                                </Badge>
                            )}
                        </div>
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
                                <MenuItem onSelect={clearHistory}>
                                    <Icon aria-hidden={true} svgPath={mdiDelete} /> Clear all chats
                                </MenuItem>
                                <MenuDivider />
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
                    {!isCTADismissed &&
                        (showVSCodeCTA ? (
                            <MarketingBlock
                                wrapperClassName="d-flex"
                                contentClassName={classNames(
                                    'flex-grow-1 d-flex flex-column justify-content-between',
                                    styles.ctaWrapper
                                )}
                            >
                                <H3 className="d-flex align-items-center mb-4">Try the VS Code Extension</H3>
                                <Text>
                                    This extension combines an LLM with the context of your code to help you generate
                                    and fix code.
                                </Text>
                                <div className="mb-2">
                                    <Link
                                        to="https://marketplace.visualstudio.com/items?itemName=sourcegraph.cody-ai"
                                        className={classNames(
                                            'd-inline-flex align-items-center text-merged',
                                            styles.ctaLink
                                        )}
                                        onClick={onDownloadVSCodeClick}
                                    >
                                        Download the VS Code Extension
                                        <Icon svgPath={mdiChevronRight} aria-hidden={true} />
                                    </Link>
                                </div>
                                <img
                                    src="https://storage.googleapis.com/sourcegraph-assets/TryCodyVSCodeExtension.png"
                                    alt="Try Cody VS Code Extension"
                                    width={666}
                                />
                                <Icon
                                    svgPath={mdiClose}
                                    aria-label="Close try Cody widget"
                                    className={classNames(styles.closeButton, 'position-absolute m-0')}
                                    onClick={onCTADismiss}
                                />
                            </MarketingBlock>
                        ) : (
                            <MarketingBlock
                                wrapperClassName="d-flex"
                                contentClassName={classNames(
                                    'flex-grow-1 d-flex flex-column justify-content-between',
                                    styles.ctaWrapper
                                )}
                            >
                                <H3 className="d-flex align-items-center mb-4">Try Cody on Public Code</H3>
                                <Text>
                                    Cody explains, generates, and translates code within specific files and
                                    repositories.
                                </Text>
                                <div className="mb-2">
                                    <Link
                                        to="https://sourcegraph.com/github.com/openai/openai-cookbook/-/blob/apps/file-q-and-a/nextjs-with-flask-server/server/answer_question.py"
                                        className={classNames(
                                            'd-inline-flex align-items-center text-merged',
                                            styles.ctaLink
                                        )}
                                        onClick={onTryOnPublicCodeClick}
                                    >
                                        Try on a file, or repository
                                        <Icon svgPath={mdiChevronRight} aria-hidden={true} />
                                    </Link>
                                </div>
                                <img
                                    src="https://storage.googleapis.com/sourcegraph-assets/TryCodyOnPublicCode.png"
                                    alt="Try Cody on Public Code"
                                    width={666}
                                />
                                <Icon
                                    svgPath={mdiClose}
                                    aria-label="Close try Cody widget"
                                    className={classNames(styles.closeButton, 'position-absolute m-0')}
                                    onClick={onCTADismiss}
                                />
                            </MarketingBlock>
                        ))}
                </div>

                {isSourcegraphApp ? (
                    <>
                        <div
                            className={classNames(
                                'col-md-9 h-100',
                                styles.chatMainWrapper,
                                showMobileHistory && styles.chatMainWrapperWithMobileHistory
                            )}
                        >
                            <div className={styles.mobileTopBar}>
                                <Button
                                    variant="icon"
                                    className={styles.mobileTopBarButton}
                                    onClick={() => setShowMobileHistory(true)}
                                >
                                    <Icon aria-hidden={true} svgPath={mdiViewList} className="mr-2" />
                                    All chats
                                </Button>
                                <div className={classNames('border-right', styles.mobileTopBarDivider)} />
                                <Button
                                    variant="icon"
                                    className={styles.mobileTopBarButton}
                                    onClick={initializeNewChat}
                                >
                                    <Icon aria-hidden={true} svgPath={mdiPlus} className="mr-2" />
                                    New chat
                                </Button>
                            </div>
                            <ChatUI codyChatStore={codyChatStore} />
                        </div>

                        {showMobileHistory && (
                            <div className={styles.mobileHistoryWrapper}>
                                <div className={styles.mobileTopBar}>
                                    <Button
                                        variant="icon"
                                        className={classNames('w-100', styles.mobileTopBarButton)}
                                        onClick={() => setShowMobileHistory(false)}
                                    >
                                        <Icon aria-hidden={true} svgPath={mdiClose} className="mr-2" />
                                        Close
                                    </Button>
                                </div>
                                <div className={styles.mobileHistory}>
                                    <HistoryList
                                        currentTranscript={transcript}
                                        transcriptHistory={transcriptHistory}
                                        truncateMessageLength={60}
                                        loadTranscriptFromHistory={loadTranscriptFromHistory}
                                        deleteHistoryItem={deleteHistoryItem}
                                    />
                                </div>
                            </div>
                        )}
                    </>
                ) : (
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
                                {showMobileHistory && (
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
                                        <Badge variant="info">Beta</Badge>
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
                            <ChatUI codyChatStore={codyChatStore} />
                        )}
                    </div>
                )}
            </div>
        </Page>
    )
}
