import { useCallback, type FC } from 'react'

import { capitalize } from 'lodash'
import { useLocation } from 'react-router-dom'

import type { AuthenticatedUser } from '@sourcegraph/shared/src/auth'
import { noOpTelemetryRecorder } from '@sourcegraph/shared/src/telemetry'
import { Link, FeedbackBadge } from '@sourcegraph/wildcard'

import { CreateGitHubAppPage } from '../../../components/gitHubApps/CreateGitHubAppPage'
import { GitHubAppDomain, GitHubAppKind } from '../../../graphql-operations'

import { useGlobalBatchChangesCodeHostConnection } from './backend'

const DEFAULT_EVENTS: string[] = []

const DEFAULT_PERMISSIONS = {
    contents: 'write',
    metadata: 'read',
}

interface BatchChangesCreateGitHubAppPageProps {
    authenticatedUser: AuthenticatedUser
    minimizedMode?: boolean
    kind: GitHubAppKind
    externalServiceURL?: string
}

export const BatchChangesCreateGitHubAppPage: FC<BatchChangesCreateGitHubAppPageProps> = ({
    minimizedMode,
    kind,
    authenticatedUser,
    externalServiceURL,
}) => {
    const location = useLocation()
    const searchParams = new URLSearchParams(location.search)
    const baseURL = externalServiceURL || searchParams.get('baseURL')

    const isGitHubAppKindCredential = kind === GitHubAppKind.USER_CREDENTIAL || kind === GitHubAppKind.SITE_CREDENTIAL

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
                const existingURL = isGitHubAppKindCredential
                    ? node.externalServiceURL
                    : node.commitSigningConfiguration?.baseURL
                if (!existingURL) {
                    return false
                }

                return new URL(existingURL).hostname === asURL.hostname
            })
            const errorMsg = `A ${
                isGitHubAppKindCredential ? 'GitHub app' : 'commit signing'
            } integration for the code host at this URL already exists.`
            return isDuplicate ? errorMsg : true
        },
        [connection, isGitHubAppKindCredential]
    )
    const pageTitle = isGitHubAppKindCredential
        ? `Create GitHub app for ${
              kind === GitHubAppKind.USER_CREDENTIAL ? authenticatedUser.username : 'Global'
          } Batch Changes credential`
        : 'Create GitHub app for commit signing'
    const defaultAppName = computeAppName(kind, authenticatedUser?.username)

    // COMMIT SIGNING apps do not need permissions to create pull request, we duplicate the
    // commit using the GraphQL request and the changeset is created with the PAT.
    const permissions = {
        ...DEFAULT_PERMISSIONS,
        ...(isGitHubAppKindCredential ? { pull_requests: 'write' } : {}),
    }
    return (
        <CreateGitHubAppPage
            minimizedMode={minimizedMode}
            authenticatedUser={authenticatedUser}
            appKind={kind}
            defaultEvents={DEFAULT_EVENTS}
            defaultPermissions={permissions}
            pageTitle={pageTitle}
            headerDescription={
                <>
                    Register a GitHub App to enable Sourcegraph{' '}
                    {isGitHubAppKindCredential ? 'create' : 'sign commits for'} Batch Change changesets on your behalf.
                    {/* TODO (@BolajiOlajide/@bahrmichael) update link here for credential github app */}
                    <Link to="/help/admin/config/batch_changes#commit-signing-for-github" className="ml-1">
                        See how GitHub App configuration works.
                    </Link>
                </>
            }
            headerAnnotation={<FeedbackBadge status="beta" feedback={{ mailto: 'support@sourcegraph.com' }} />}
            appDomain={GitHubAppDomain.BATCHES}
            defaultAppName={defaultAppName}
            baseURL={baseURL?.length ? baseURL : undefined}
            validateURL={validateURL}
            telemetryRecorder={noOpTelemetryRecorder}
        />
    )
}

const computeAppName = (kind: GitHubAppKind, username?: string): string => {
    switch (kind) {
        case GitHubAppKind.COMMIT_SIGNING: {
            return 'Sourcegraph Commit Signing'
        }

        case GitHubAppKind.USER_CREDENTIAL: {
            return `${capitalize(username)}'s Batch Changes GitHub App`
        }

        default: {
            return 'Batch Changes GitHub App'
        }
    }
}
