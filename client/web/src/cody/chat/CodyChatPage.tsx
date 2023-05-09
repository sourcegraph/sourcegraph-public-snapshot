import React, { useEffect } from 'react'

import {
    mdiCogOutline,
    mdiDelete,
    mdiDotsVertical,
    mdiOpenInNew,
    mdiPlus,
    mdiMicrosoftVisualStudioCode,
    mdiChevronRight,
} from '@mdi/js'
import classNames from 'classnames'

import { AuthenticatedUser } from '@sourcegraph/shared/src/auth'
import {
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
    ButtonLink,
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

import { CodyPageIcon } from './CodyPageIcon'

import styles from './CodyChatPage.module.scss'

interface CodyChatPageProps {
    authenticatedUser: AuthenticatedUser | null
}

export const CodyChatPage: React.FunctionComponent<CodyChatPageProps> = ({ authenticatedUser }) => {
    const { reset, clearHistory } = useChatStore({ codebase: '' })
    const [enabled] = useFeatureFlag('cody-web-chat')
    const onInstallClick = (): void => eventLogger.log(EventName.TRY_CODY_VSCODE)
    const onMarketplaceClick = (): void => eventLogger.log(EventName.TRY_CODY_MARKETPLACE)

    useEffect(() => {
        eventLogger.logPageView('CodyChat')
    }, [])

    if (!enabled) {
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
                    <PageHeader.Breadcrumb icon={CodyPageIcon}> Cody Chat</PageHeader.Breadcrumb>
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
                    <MarketingBlock
                        wrapperClassName="d-flex"
                        contentClassName="flex-grow-1 d-flex flex-column justify-content-between p-3 bg-white"
                    >
                        <H3 className="d-flex align-items-center">
                            <Icon
                                svgPath={mdiMicrosoftVisualStudioCode}
                                aria-hidden={true}
                                className="mr-1 text-primary"
                                size="md"
                            />
                            Download for VS Code
                        </H3>
                        <Text>Get the power of Cody in your editor.</Text>
                        <div className="mb-2">
                            <ButtonLink
                                to="vscode:extension/sourcegraph.cody-ai"
                                variant="merged"
                                className="d-inline-flex align-items-center"
                                onClick={onInstallClick}
                            >
                                Install Cody for VS Code <Icon svgPath={mdiChevronRight} aria-hidden={true} size="md" />
                            </ButtonLink>
                        </div>
                        <Text
                            size="small"
                            as={Link}
                            to="https://marketplace.visualstudio.com/items?itemName=sourcegraph.cody-ai"
                            target="_blank"
                            rel="noopener"
                            onClick={onMarketplaceClick}
                        >
                            or download on the VS Code marketplace
                        </Text>
                    </MarketingBlock>
                </div>

                <div className={classNames('d-flex flex-column col-sm-9 h-100', styles.chatMainWrapper)}>
                    <ChatUI />
                </div>
            </div>
        </Page>
    )
}
