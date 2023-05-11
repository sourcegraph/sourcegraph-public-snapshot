import React, { useEffect, useState } from 'react'

import { mdiClose, mdiCogOutline, mdiDelete, mdiDotsVertical, mdiOpenInNew, mdiPlus, mdiChevronRight } from '@mdi/js'
import classNames from 'classnames'

import { AuthenticatedUser } from '@sourcegraph/shared/src/auth'
import { useTemporarySetting } from '@sourcegraph/shared/src/settings/temporary'
import {
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
} from '@sourcegraph/wildcard'

import { MarketingBlock } from '../../components/MarketingBlock'
import { Page } from '../../components/Page'
import { PageTitle } from '../../components/PageTitle'
import { useFeatureFlag } from '../../featureFlags/useFeatureFlag'
import { eventLogger } from '../../tracking/eventLogger'
import { EventName } from '../../util/constants'
import { ChatUI } from '../components/ChatUI'
import { HistoryList } from '../components/HistoryList'
import { useChatStore } from '../stores/chat'

import { CodyColorIcon } from './CodyPageIcon'

import styles from './CodyChatPage.module.scss'

interface CodyChatPageProps {
    authenticatedUser: AuthenticatedUser | null
}

const onDownloadVSCodeClick = (): void => eventLogger.log(EventName.CODY_CHAT_DOWNLOAD_VSCODE)
const onTryOnPublicCodeClick = (): void => eventLogger.log(EventName.CODY_CHAT_TRY_ON_PUBLIC_CODE)

export const CodyChatPage: React.FunctionComponent<CodyChatPageProps> = ({ authenticatedUser }) => {
    const { reset, clearHistory } = useChatStore({ codebase: '' })
    const [enabled] = useFeatureFlag('cody-web-chat')
    const [showVSCodeCTA] = useState<boolean>(Math.random() < 0.5 || true)
    const [isCTADismissed = true, setIsCTADismissed] = useTemporarySetting('cody.chatPageCta.dismissed', false)
    const onCTADismiss = (): void => setIsCTADismissed(true)

    useEffect(() => {
        eventLogger.logPageView('CodyChat')
    }, [])

    if (!enabled || !authenticatedUser) {
        return null
    }

    return (
        <Page className="overflow-hidden">
            <PageTitle title="Cody AI Chat" />
            <PageHeader
                actions={
                    <div className="d-flex">
                        <Button variant="primary" onClick={reset}>
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
                className="mb-3"
            >
                <PageHeader.Heading as="h2" styleAs="h1">
                    <PageHeader.Breadcrumb icon={CodyColorIcon}>
                        <div className="d-inline-flex align-items-center">
                            Cody Chat
                            <Badge variant="info" className="ml-2">
                                Beta
                            </Badge>
                        </div>
                    </PageHeader.Breadcrumb>
                </PageHeader.Heading>
            </PageHeader>

            {/* Page content */}
            <div className={classNames('row mb-5', styles.pageWrapper)}>
                <div className="d-flex flex-column col-sm-3 h-100">
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
                        <HistoryList trucateMessageLenght={60} />
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

                <div className={classNames('d-flex flex-column col-sm-9 h-100', styles.chatMainWrapper)}>
                    <ChatUI />
                </div>
            </div>
        </Page>
    )
}
