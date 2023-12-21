import React, { useEffect } from 'react'

import { useApolloClient } from '@apollo/client'
import { useLocation } from 'react-router-dom'

import { logger } from '@sourcegraph/common'
import type { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Alert, H4, Link, PageHeader } from '@sourcegraph/wildcard'

import { PageTitle } from '../components/PageTitle'
import { refreshSiteFlags } from '../site/backend'

import { SiteAdminRepositoriesContainer } from './SiteAdminRepositoriesContainer'

interface Props extends TelemetryProps {}

/** A page displaying the repositories on this site */
export const SiteAdminRepositoriesPage: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
    telemetryService,
}) => {
    const location = useLocation()

    useEffect(() => {
        telemetryService.logPageView('SiteAdminRepos')
    }, [telemetryService])

    // Refresh global alert about enabling repositories when the user visits & navigates away from this page.
    const client = useApolloClient()
    useEffect(() => {
        refreshSiteFlags(client).then(null, error => logger.error(error))
        return () => {
            refreshSiteFlags(client).then(null, error => logger.error(error))
        }
    }, [client])

    const showRepositoriesAddedBanner = new URLSearchParams(location.search).has('repositoriesUpdated')

    const licenseInfo = window.context.licenseInfo

    return (
        <div className="site-admin-repositories-page">
            <PageTitle title="Repositories - Admin" />
            {showRepositoriesAddedBanner && (
                <Alert variant="success" as="p">
                    Syncing repositories. It may take a few moments to clone and index each repository. Repository
                    statuses are displayed below.
                </Alert>
            )}
            <PageHeader
                path={[{ text: 'Repositories' }]}
                headingElement="h2"
                description={
                    <>
                        Repositories are synced from connected{' '}
                        <Link
                            to="/site-admin/external-services"
                            data-testid="test-repositories-code-host-connections-link"
                        >
                            code host connections
                        </Link>
                        .
                    </>
                }
                className="mb-3"
            />

            {licenseInfo && (licenseInfo.codeScaleCloseToLimit || licenseInfo.codeScaleExceededLimit) && (
                <Alert variant={licenseInfo.codeScaleExceededLimit ? 'danger' : 'warning'}>
                    <H4>
                        {licenseInfo.codeScaleExceededLimit ? (
                            <>You've used all 100GiB of storage</>
                        ) : (
                            <>Your Sourcegraph is almost full</>
                        )}
                    </H4>
                    {licenseInfo.codeScaleExceededLimit ? <>You're about to reach the 100GiB storage limit. </> : <></>}
                    Upgrade to <Link to="https://sourcegraph.com/pricing">Sourcegraph Enterprise</Link> for unlimited
                    storage for your code.
                </Alert>
            )}
            <SiteAdminRepositoriesContainer />
        </div>
    )
}
