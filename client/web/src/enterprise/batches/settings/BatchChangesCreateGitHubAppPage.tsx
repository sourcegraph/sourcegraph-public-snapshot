import { useCallback } from 'react'

import { useLocation } from 'react-router-dom'

import { FeedbackBadge, Link } from '@sourcegraph/wildcard'

import { CreateGitHubAppPage } from '../../../components/gitHubApps/CreateGitHubAppPage'
import { GitHubAppDomain } from '../../../graphql-operations'

import { useGlobalBatchChangesCodeHostConnection } from './backend'

const DEFAULT_EVENTS: string[] = []

const DEFAULT_PERMISSIONS = {
    contents: 'write',
    metadata: 'read',
}

export const BatchChangesCreateGitHubAppPage: React.FunctionComponent = () => {
    const location = useLocation()
    const baseURL = new URLSearchParams(location.search).get('baseURL')

    const { connection } = useGlobalBatchChangesCodeHostConnection()
    // validateURL compares a provided URL against the URLs of existing commit signing
    // integrations for code hosts, to ensure the new GH App is for a unique code host.
    const validateURL = useCallback(
        (url: string): true | string => {
            if (!connection) {
                // We don't have a connection yet, so we can't validate that the URL is
                // unique, but we shouldn't block on that.
                return true
            }
            // The default validator already checks that the URL is a valid URL, so we
            // assume this call will succeed.
            const asURL = new URL(url)
            const isDuplicate = connection.nodes.some(node => {
                const existingURL = node.commitSigningConfiguration?.baseURL
                if (!existingURL) {
                    return false
                }

                return new URL(existingURL).hostname === asURL.hostname
            })
            return isDuplicate ? 'A commit signing integration for the code host at this URL already exists.' : true
        },
        [connection]
    )
    return (
        <CreateGitHubAppPage
            defaultEvents={DEFAULT_EVENTS}
            defaultPermissions={DEFAULT_PERMISSIONS}
            pageTitle="Create GitHub App for commit signing"
            headerDescription={
                <>
                    Register a GitHub App to enable Sourcegraph to sign commits for Batch Change changesets on your
                    behalf.
                    <Link to="/help/admin/config/batch_changes#commit-signing-for-github" className="ml-1">
                        See how GitHub App configuration works.
                    </Link>
                </>
            }
            headerAnnotation={<FeedbackBadge status="beta" feedback={{ mailto: 'support@sourcegraph.com' }} />}
            appDomain={GitHubAppDomain.BATCHES}
            defaultAppName="Sourcegraph Commit Signing"
            baseURL={baseURL?.length ? baseURL : undefined}
            validateURL={validateURL}
        />
    )
}
