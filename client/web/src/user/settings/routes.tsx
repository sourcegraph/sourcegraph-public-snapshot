import type { FC } from 'react'

import { Navigate } from 'react-router-dom'

import type { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import type { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryV2Props } from '@sourcegraph/shared/src/telemetry'
import type { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { useIsLightTheme } from '@sourcegraph/shared/src/theme'
import { lazyComponent } from '@sourcegraph/shared/src/util/lazyComponent'
import { Text } from '@sourcegraph/wildcard'

import type { AuthenticatedUser } from '../../auth'
import type { UserSettingsAreaUserFields } from '../../graphql-operations'
import { SiteAdminAlert } from '../../site-admin/SiteAdminAlert'

import type { UserSettingsAreaRoute } from './UserSettingsArea'

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
