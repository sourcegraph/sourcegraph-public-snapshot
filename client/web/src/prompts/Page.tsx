import type { FunctionComponent, ReactNode } from 'react'

import { mdiMagnify } from '@mdi/js'
import { useLocation } from 'react-router-dom'

import { PageHeader } from '@sourcegraph/wildcard'

import { Page } from '../components/Page'
import { PageTitle } from '../components/PageTitle'
import type { PromptFields } from '../graphql-operations'
import { PageRoutes } from '../routes.constants'

import { urlToPromptsList } from './ListPage'
import { PromptIcon } from './PromptIcon'

/**
 * The template for a prompt page.
 */
export const PromptPage: FunctionComponent<{
    title?: string
    actions?: ReactNode
    breadcrumbsNamespace?: PromptFields['owner']
    breadcrumbs?: ReactNode
    children: ReactNode
}> = ({ title, actions, breadcrumbsNamespace, breadcrumbs, children }) => {
    const location = useLocation()
    const isRootPage = location.pathname === PageRoutes.Prompts

    return (
        <Page className="w-100">
            {title && <PageTitle title={title} />}
            <PageHeader actions={actions} className="mb-3">
                <PageHeader.Heading as="h2" styleAs="h1" className="mb-1">
                    <PageHeader.Breadcrumb icon={mdiMagnify} to="/search" aria-label="Code Search" />
                    <PageHeader.Breadcrumb icon={PromptIcon} to={isRootPage ? undefined : PageRoutes.Prompts}>
                        Prompt Library
                    </PageHeader.Breadcrumb>
                    {breadcrumbsNamespace && (
                        <PageHeader.Breadcrumb to={urlToPromptsList(breadcrumbsNamespace.id)}>
                            {breadcrumbsNamespace.namespaceName}
                        </PageHeader.Breadcrumb>
                    )}
                    {breadcrumbs}
                </PageHeader.Heading>
            </PageHeader>
            {children}
            <div className="pb-4" />
        </Page>
    )
}
