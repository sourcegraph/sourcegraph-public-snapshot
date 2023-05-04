import { FC } from 'react'

import { Routes, Route } from 'react-router-dom'
import { SiteExternalServiceConfigResult, SiteExternalServiceConfigVariables } from 'src/graphql-operations'

import { useQuery } from '@sourcegraph/http-client'
import { AuthenticatedUser } from '@sourcegraph/shared/src/auth'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { lazyComponent } from '@sourcegraph/shared/src/util/lazyComponent'
import { LoadingSpinner, ErrorAlert } from '@sourcegraph/wildcard'

import { SITE_EXTERNAL_SERVICE_CONFIG } from './backend'

const AddGitHubAppPage = lazyComponent(() => import('../components/gitHubApps/AddGitHubAppPage'), 'AddGitHubAppPage')
const GitHubAppPage = lazyComponent(() => import('../components/gitHubApps/GitHubAppPage'), 'GitHubAppPage')
const GitHubAppsPage = lazyComponent(() => import('../components/gitHubApps/GitHubAppsPage'), 'GitHubAppsPage')

interface Props extends TelemetryProps, PlatformContextProps {
    authenticatedUser: AuthenticatedUser
    isSourcegraphApp: boolean
}

export const SiteAdminGitHubAppsArea: FC<Props> = props => {
    const { data, error, loading } = useQuery<SiteExternalServiceConfigResult, SiteExternalServiceConfigVariables>(
        SITE_EXTERNAL_SERVICE_CONFIG,
        {}
    )

    if (error && !loading) {
        return <ErrorAlert error={error} />
    }

    if (loading && !error) {
        return <LoadingSpinner />
    }

    if (!data) {
        return null
    }

    return (
        <Routes>
            <Route index={true} element={<GitHubAppsPage />} />

            <Route path="new" element={<AddGitHubAppPage {...props} />} />
            <Route
                path=":appID"
                element={
                    <GitHubAppPage
                        {...props}
                        externalServicesFromFile={data?.site?.externalServicesFromFile}
                        allowEditExternalServicesWithFile={data?.site?.allowEditExternalServicesWithFile}
                    />
                }
            />
        </Routes>
    )
}
