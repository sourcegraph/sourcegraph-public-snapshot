import React, { Suspense, useMemo } from 'react'

import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import MapSearchIcon from 'mdi-react/MapSearchIcon'
import { Route, Routes, useLocation, useNavigate, useParams } from 'react-router-dom'

import { gql, useQuery } from '@sourcegraph/http-client'
import type { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import type { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import type { TelemetryV2Props } from '@sourcegraph/shared/src/telemetry'
import type { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ErrorMessage, LoadingSpinner } from '@sourcegraph/wildcard'

import type { AuthenticatedUser } from '../../auth'
import type { BatchChangesProps } from '../../batches'
import type { BreadcrumbSetters, BreadcrumbsProps } from '../../components/Breadcrumbs'
import { RouteError } from '../../components/ErrorBoundary'
import { HeroPage } from '../../components/HeroPage'
import { Page } from '../../components/Page'
import type { OrgAreaOrganizationFields } from '../../graphql-operations'
import type { NamespaceProps } from '../../namespaces'
import type { RouteV6Descriptor } from '../../util/contributions'
import type { OrgSettingsAreaRoute } from '../settings/OrgSettingsArea'
import type { OrgSettingsSidebarItems } from '../settings/OrgSettingsSidebar'

import { type OrgAreaHeaderNavItem, OrgHeader } from './OrgHeader'
import { OrgInvitationPageLegacy } from './OrgInvitationPageLegacy'

export interface OrgAreaRoute extends RouteV6Descriptor<OrgAreaRouteContext> {
    /** When true, the header is not rendered and the component is not wrapped in a container. */
    fullPage?: boolean
}

export interface OrgAreaProps
    extends PlatformContextProps,
        SettingsCascadeProps,
        TelemetryProps,
        BreadcrumbsProps,
        BreadcrumbSetters,
        BatchChangesProps {
    orgAreaRoutes: readonly OrgAreaRoute[]
    orgAreaHeaderNavItems: readonly OrgAreaHeaderNavItem[]
    orgSettingsSideBarItems: OrgSettingsSidebarItems
    orgSettingsAreaRoutes: readonly OrgSettingsAreaRoute[]

    /**
     * The currently authenticated user.
     */
    authenticatedUser: AuthenticatedUser
}

const ORGANIZATION_QUERY = gql`
    query Organization($name: String!) {
        organization(name: $name) {
            ...OrgAreaOrganizationFields
        }
    }

    fragment OrgAreaOrganizationFields on Org {
        __typename
        id
        name
        displayName
        url
        settingsURL
        viewerPendingInvitation {
            id
            sender {
                username
                displayName
                avatarURL
                createdAt
            }
            respondURL
        }
        viewerIsMember
        viewerCanAdminister
        createdAt
    }
`

const NotFoundPage: React.FunctionComponent<React.PropsWithChildren<unknown>> = () => (
    <HeroPage icon={MapSearchIcon} title="404: Not Found" subtitle="Sorry, the requested organization was not found." />
)

/**
 * Properties passed to all page components in the org area.
 */
export interface OrgAreaRouteContext
    extends PlatformContextProps,
        SettingsCascadeProps,
        TelemetryProps,
        TelemetryV2Props,
        NamespaceProps,
        BreadcrumbsProps,
        BreadcrumbSetters,
        BatchChangesProps {
    /** The org that is the subject of the page. */
    org: OrgAreaOrganizationFields

    /** Called when the organization is updated and must be reloaded. */
    onOrganizationUpdate: () => void

    /** The currently authenticated user. */
    authenticatedUser: AuthenticatedUser

    orgSettingsSideBarItems: OrgSettingsSidebarItems
    orgSettingsAreaRoutes: readonly OrgSettingsAreaRoute[]
}

export const OrgArea: React.FunctionComponent<OrgAreaProps> = ({
    orgAreaRoutes,
    orgAreaHeaderNavItems,
    orgSettingsSideBarItems,
    orgSettingsAreaRoutes,
    authenticatedUser,
    useBreadcrumb,
    ...props
}) => {
    const { orgName } = useParams<{ orgName: string }>()
    if (!orgName) {
        throw new Error('orgName is required')
    }

    const navigate = useNavigate()
    const location = useLocation()

    const { data, error, loading, refetch } = useQuery(ORGANIZATION_QUERY, {
        variables: { name: orgName },
    })

    const childBreadcrumbSetters = useBreadcrumb(
        useMemo(
            () =>
                data?.organization
                    ? {
                          key: 'OrgArea',
                          link: { to: data.organization.url, label: data.organization.name },
                      }
                    : null,
            [data]
        )
    )

    if (loading && !data) {
        return <LoadingSpinner className="m-2" />
    }

    if (error) {
        return <HeroPage icon={AlertCircleIcon} title="Error" subtitle={<ErrorMessage error={error} />} />
    }

    if (!data?.organization) {
        return <NotFoundPage />
    }

    const context: OrgAreaRouteContext = {
        authenticatedUser,
        org: data.organization,
        onOrganizationUpdate: refetch,
        platformContext: props.platformContext,
        settingsCascade: props.settingsCascade,
        namespace: data.organization,
        telemetryService: props.telemetryService,
        telemetryRecorder: props.platformContext.telemetryRecorder,
        batchChangesEnabled: props.batchChangesEnabled,
        batchChangesExecutionEnabled: props.batchChangesExecutionEnabled,
        batchChangesWebhookLogsEnabled: props.batchChangesWebhookLogsEnabled,
        breadcrumbs: props.breadcrumbs,
        ...childBreadcrumbSetters,
        orgSettingsAreaRoutes,
        orgSettingsSideBarItems,
    }

    const handleRespondToInvitation = (accepted: boolean): void => {
        if (!accepted) {
            navigate('/user/settings')
            return
        }
        refetch()
    }

    if (location.pathname === `/organizations/${orgName}/invitation`) {
        return <OrgInvitationPageLegacy {...context} onDidRespondToInvitation={handleRespondToInvitation} />
    }

    return (
        <Suspense fallback={<LoadingSpinner className="m-2" />}>
            <Routes>
                {orgAreaRoutes.map(
                    ({ path, render, condition = () => true, fullPage }) =>
                        condition(context) && (
                            <Route
                                path={path}
                                key="hardcoded-key"
                                errorElement={<RouteError />}
                                element={
                                    fullPage ? (
                                        render(context)
                                    ) : (
                                        <Page className="org-area">
                                            <OrgHeader
                                                {...props}
                                                {...context}
                                                navItems={orgAreaHeaderNavItems}
                                                className="mb-3"
                                            />
                                            <div className="container">{render(context)}</div>
                                        </Page>
                                    )
                                }
                            />
                        )
                )}
                <Route path="*" element={<NotFoundPage />} />
            </Routes>
        </Suspense>
    )
}
