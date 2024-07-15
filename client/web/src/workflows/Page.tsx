import type { FunctionComponent, ReactNode } from 'react'

import { mdiMagnify } from '@mdi/js'
import { useLocation } from 'react-router-dom'

import { PageHeader } from '@sourcegraph/wildcard'

import { Page } from '../components/Page'
import { PageTitle } from '../components/PageTitle'
import { type WorkflowFields } from '../graphql-operations'
import { PageRoutes } from '../routes.constants'

import { urlToWorkflowsList } from './ListPage'
import { WorkflowIcon } from './WorkflowIcon'

/**
 * The template for a workflow page.
 */
export const WorkflowPage: FunctionComponent<{
    title?: string
    actions?: ReactNode
    breadcrumbsNamespace?: WorkflowFields['owner']
    breadcrumbs?: ReactNode
    children: ReactNode
}> = ({ title, actions, breadcrumbsNamespace, breadcrumbs, children }) => {
    const location = useLocation()
    const isRootPage = location.pathname === PageRoutes.Workflows

    return (
        <Page className="w-100">
            {title && <PageTitle title={title} />}
            <PageHeader actions={actions} className="mb-3">
                <PageHeader.Heading as="h2" styleAs="h1" className="mb-1">
                    <PageHeader.Breadcrumb icon={mdiMagnify} to="/search" aria-label="Code Search" />
                    <PageHeader.Breadcrumb
                        icon={WorkflowIcon}
                        to={isRootPage ? undefined : PageRoutes.Workflows}
                    >
                        Workflows
                    </PageHeader.Breadcrumb>
                    {breadcrumbsNamespace && (
                        <PageHeader.Breadcrumb to={urlToWorkflowsList(breadcrumbsNamespace.id)}>
                            {breadcrumbsNamespace.namespaceName}
                        </PageHeader.Breadcrumb>
                    )}
                    {breadcrumbs}
                </PageHeader.Heading>
            </PageHeader>
            {children}
        </Page>
    )
}
