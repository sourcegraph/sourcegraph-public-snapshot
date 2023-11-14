import type { FC } from 'react'

import { Navigate } from 'react-router-dom'

import type { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import type { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import type { TelemetryV2Props } from '@sourcegraph/shared/src/telemetry'
import type { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { useIsLightTheme } from '@sourcegraph/shared/src/theme'
import { lazyComponent } from '@sourcegraph/shared/src/util/lazyComponent'
import { Text } from '@sourcegraph/wildcard'

import type { AuthenticatedUser } from '../../auth'
import { canWriteBatchChanges } from '../../batches/utils'
import { SHOW_BUSINESS_FEATURES } from '../../enterprise/dotcom/productSubscriptions/features'
import type { ExecutorsUserAreaProps } from '../../enterprise/executors/ExecutorsUserArea'
import type { UserEventLogsPageProps } from '../../enterprise/user/settings/UserEventLogsPage'
import type { UserSettingsAreaUserFields } from '../../graphql-operations'
import { SiteAdminAlert } from '../../site-admin/SiteAdminAlert'

import type { UserSettingsAreaRoute } from './UserSettingsArea'

const ExecutorsUserArea = lazyComponent<ExecutorsUserAreaProps, 'ExecutorsUserArea'>(
    () => import('../../enterprise/executors/ExecutorsUserArea'),
    'ExecutorsUserArea'
)

const SettingsArea = lazyComponent(() => import('../../settings/SettingsArea'), 'SettingsArea')

const UserSettingsSecurityPage = lazyComponent(
    () => import('./auth/UserSettingsSecurityPage'),
    'UserSettingsSecurityPage'
)

export const userSettingsAreaRoutes: readonly UserSettingsAreaRoute[] = [
    {
        path: '',
        render: props => <UserSettingAreaIndexPage {...props} />,
    },
    {
        path: 'profile',
        render: lazyComponent(() => import('./profile/UserSettingsProfilePage'), 'UserSettingsProfilePage'),
    },
    {
        path: 'password',
        render: () => <Navigate to="../security" replace={true} />,
    },
    {
        path: 'quota',
        render: lazyComponent(() => import('./quota/UserQuotaProfilePage'), 'UserQuotaProfilePage'),
        condition: ({ authenticatedUser }) => authenticatedUser.siteAdmin,
    },
    {
        path: 'emails',
        render: lazyComponent(() => import('./emails/UserSettingsEmailsPage'), 'UserSettingsEmailsPage'),
    },
    {
        path: 'tokens/*',
        render: lazyComponent(() => import('./accessTokens/UserSettingsTokensArea'), 'UserSettingsTokensArea'),
        condition: () => window.context.accessTokensAllow !== 'none',
    },
    // future GA Cloud routes
    {
        path: 'security',
        render: props => <UserSettingsSecurityPage {...props} context={window.context} />,
    },
    {
        path: 'product-research',
        render: lazyComponent(() => import('./research/ProductResearch'), 'ProductResearchPage'),
        condition: () => window.context.productResearchPageEnabled,
    },
    {
        path: 'about-organizations',
        render: lazyComponent(() => import('./aboutOrganization/AboutOrganizationPage'), 'AboutOrganizationPage'),
    },
    {
        path: 'permissions',
        render: lazyComponent(
            () => import('../../enterprise/user/settings/auth/UserSettingsPermissionsPage'),
            'UserSettingsPermissionsPage'
        ),
    },
    {
        path: 'event-log',
        render: lazyComponent<UserEventLogsPageProps, 'UserEventLogsPage'>(
            () => import('../../enterprise/user/settings/UserEventLogsPage'),
            'UserEventLogsPage'
        ),
    },
    {
        path: 'executors/*',
        render: props => <ExecutorsUserArea {...props} namespaceID={props.user.id} />,
        condition: ({ batchChangesEnabled, user: { viewerCanAdminister }, authenticatedUser }) =>
            batchChangesEnabled && viewerCanAdminister && canWriteBatchChanges(authenticatedUser),
    },
    {
        path: 'batch-changes',
        render: lazyComponent(
            () => import('../../enterprise/batches/settings/BatchChangesSettingsArea'),
            'BatchChangesSettingsArea'
        ),
        condition: ({ batchChangesEnabled, user: { viewerCanAdminister }, authenticatedUser }) =>
            batchChangesEnabled && viewerCanAdminister && canWriteBatchChanges(authenticatedUser),
    },
    {
        path: 'subscriptions/:subscriptionUUID',
        render: lazyComponent(
            () => import('../../enterprise/user/productSubscriptions/UserSubscriptionsProductSubscriptionPage'),
            'UserSubscriptionsProductSubscriptionPage'
        ),
        condition: () => SHOW_BUSINESS_FEATURES,
    },
    {
        path: 'subscriptions',
        render: lazyComponent(
            () => import('../../enterprise/user/productSubscriptions/UserSubscriptionsProductSubscriptionsPage'),
            'UserSubscriptionsProductSubscriptionsPage'
        ),
        condition: () => SHOW_BUSINESS_FEATURES,
    },
]

interface UserSettingAreaIndexPageProps
    extends PlatformContextProps,
        SettingsCascadeProps,
        TelemetryProps,
        TelemetryV2Props {
    isSourcegraphDotCom: boolean
    authenticatedUser: AuthenticatedUser
    user: UserSettingsAreaUserFields
    extraHeader?: JSX.Element
    className?: string
}

const UserSettingAreaIndexPage: FC<UserSettingAreaIndexPageProps> = props => {
    const { isSourcegraphDotCom, authenticatedUser, user } = props
    const isLightTheme = useIsLightTheme()

    if (isSourcegraphDotCom && authenticatedUser && user.id !== authenticatedUser.id) {
        return (
            <SiteAdminAlert className="sidebar__alert" variant="danger">
                Only the user may access their individual settings.
            </SiteAdminAlert>
        )
    }
    return (
        <SettingsArea
            {...props}
            subject={props.user}
            isLightTheme={isLightTheme}
            extraHeader={
                <>
                    {authenticatedUser && user.id !== authenticatedUser.id && (
                        <SiteAdminAlert className="sidebar__alert">
                            Viewing settings for <strong>{user.username}</strong>
                        </SiteAdminAlert>
                    )}
                    <Text>User settings override global and organization settings.</Text>
                </>
            }
        />
    )
}
