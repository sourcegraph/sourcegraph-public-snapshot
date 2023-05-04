import React from 'react'

import { mdiCogOutline, mdiDelete, mdiDotsVertical, mdiOpenInNew, mdiPlus } from '@mdi/js'
import classNames from 'classnames'

import { AuthenticatedUser } from '@sourcegraph/shared/src/auth'
import { TelemetryService } from '@sourcegraph/shared/src/telemetry/telemetryService'
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
} from '@sourcegraph/wildcard'

import { Page } from '../../components/Page'
import { PageTitle } from '../../components/PageTitle'
import { useFeatureFlag } from '../../featureFlags/useFeatureFlag'
import { ChatUI } from '../components/ChatUI'
import { HistoryList } from '../components/HistoryList'
import { useChatStore } from '../stores/chat'

import { CodyPageIcon } from './CodyPageIcon'

import styles from './CodyChatPage.module.scss'

interface CodyChatPageProps {
    authenticatedUser: AuthenticatedUser | null
    telemetryService: TelemetryService
}

export const CodyChatPage: React.FunctionComponent<CodyChatPageProps> = ({ authenticatedUser }) => {
    const { reset, clearHistory } = useChatStore({ codebase: '' })
    const [enabled] = useFeatureFlag('cody-web-chat')

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
                    <div className={classNames('h-100', styles.sidebar)}>
                        <HistoryList trucateMessageLenght={60} />
                    </div>
                </div>

                <div className={classNames('d-flex flex-column col-sm-9 h-100', styles.chatMainWrapper)}>
                    <ChatUI />
                </div>
            </div>
        </Page>
    )
}
