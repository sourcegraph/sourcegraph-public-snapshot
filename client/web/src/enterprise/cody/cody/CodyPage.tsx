import React from 'react'

import { mdiPlus } from '@mdi/js'
import classNames from 'classnames'

import { AuthenticatedUser } from '@sourcegraph/shared/src/auth'
import { TelemetryService } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Button, Icon, PageHeader } from '@sourcegraph/wildcard'

import { ChatHistory } from '../../../cody/ChatHistory'
import { CodyChat } from '../../../cody/CodyChat'
import { Page } from '../../../components/Page'
import { PageTitle } from '../../../components/PageTitle'
import { useChatStore, useChatStoreState } from '../../../stores/codyChat'
import { useCodySidebarStore } from '../../../stores/codySidebar'

import { CodyPageIcon } from './CodyPageIcon'

import styles from './CodyPage.module.scss'

interface CodePageProps {
    authenticatedUser: AuthenticatedUser | null
    telemetryService: TelemetryService
}

export const CodyPage: React.FunctionComponent<CodePageProps> = ({ authenticatedUser }) => {
    const { setIsOpen: setIsCodySidebarOpen } = useCodySidebarStore()
    // TODO: This hook call is used to initialize the chat store with the right repo name.
    useChatStore({ codebase: '', setIsCodySidebarOpen })
    const { reset, clearHistory, loadTranscriptFromHistory, transcriptHistory } = useChatStoreState()

    return (
        <Page className="overflow-hidden">
            <PageTitle title="Cody AI" />
            <PageHeader
                actions={
                    <Button variant="primary" onClick={reset}>
                        <Icon aria-hidden={true} svgPath={mdiPlus} /> New chat
                    </Button>
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
                    <PageHeader.Breadcrumb icon={CodyPageIcon}>Cody AI</PageHeader.Breadcrumb>
                </PageHeader.Heading>
            </PageHeader>

            {/* Page content */}
            <div className={classNames('row mb-5', styles.pageWrapper)}>
                <div className={classNames('d-flex flex-column col-sm-3 h-100', styles.sidebar)}>
                    <h4>
                        <b>Conversations</b>
                    </h4>
                    <ChatHistory
                        transcriptHistory={transcriptHistory}
                        loadTranscript={loadTranscriptFromHistory}
                        closeHistory={(): void => {}}
                        clearHistory={clearHistory}
                        showHeader={false}
                        itemBodyClass={styles.historyItemBody}
                        trucateMessageLenght={60}
                    />
                </div>

                <div className={classNames('d-flex flex-column col-sm-9 h-100')}>
                    <CodyChat showHeader={false} chatWrapperClass={styles.chatMainWrapper} />
                </div>
            </div>
        </Page>
    )
}
