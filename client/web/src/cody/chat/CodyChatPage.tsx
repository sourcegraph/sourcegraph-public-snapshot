import type { FC } from 'react'

import { Navigate } from 'react-router-dom'

import { CodyWebPanelProvider } from '@sourcegraph/cody-web'
import type { AuthenticatedUser } from '@sourcegraph/shared/src/auth'
import { ButtonLink, PageHeader, ProductStatusBadge } from '@sourcegraph/wildcard'

import { withAuthenticatedUser } from '../../auth/withAuthenticatedUser'
import { Page } from '../../components/Page'
import { PageTitle } from '../../components/PageTitle'
import { PageRoutes } from '../../routes.constants'
import { getTelemetrySourceClient } from '../../telemetry'
import { CodyProRoutes } from '../codyProRoutes'
import { ChatUi } from '../components/ChatUi'

import { CodyColorIcon } from './CodyPageIcon'

import styles from './CodyChatPage.module.scss'

interface CodyChatPageProps {
    isSourcegraphDotCom: boolean
    authenticatedUser: AuthenticatedUser
}

const AuthenticatedCodyChatPage: FC<CodyChatPageProps> = props => {
    const { isSourcegraphDotCom, authenticatedUser } = props

    if (!authenticatedUser) {
        return null
    }

    return (
        <Page className={styles.root}>
            <PageTitle title="Cody Chat" />

            <CodyPageHeader isSourcegraphDotCom={isSourcegraphDotCom} className={styles.pageHeader} />

            <div className={styles.chatContainer}>
                <CodyWebPanelProvider
                    accessToken=""
                    serverEndpoint={window.location.origin}
                    customHeaders={window.context.xhrHeaders}
                    telemetryClientName={getTelemetrySourceClient()}
                >
                    <ChatUi className={styles.chat} />
                </CodyWebPanelProvider>
            </div>
        </Page>
    )
}

export const CodyChatPage = withAuthenticatedUser(AuthenticatedCodyChatPage)

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
                        <ProductStatusBadge status="beta" className="ml-2" />
                    </div>
                </PageHeader.Breadcrumb>
            </PageHeader.Heading>
        </PageHeader>
    )
}
