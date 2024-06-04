import { useCallback, type FC } from 'react'

import { useLocation } from 'react-router-dom'

import type { AuthenticatedUser } from '@sourcegraph/shared/src/auth'
import { noOpTelemetryRecorder } from '@sourcegraph/shared/src/telemetry'
import { FeedbackBadge, Link } from '@sourcegraph/wildcard'

import { CreateGitHubAppPage } from '../../../components/gitHubApps/CreateGitHubAppPage'
import { GitHubAppDomain, GitHubAppKind } from '../../../graphql-operations'

import { useGlobalBatchChangesCodeHostConnection } from './backend'

const DEFAULT_EVENTS: string[] = []

const DEFAULT_PERMISSIONS = {
    contents: 'write',
    metadata: 'read',
}

const computeGitHubAppKind = (kind: string): GitHubAppKind => {
    if (kind === 'USER_CREDENTIAL') {
        return GitHubAppKind.USER_CREDENTIAL
    }

    // We default to commit signing always, since this was initially built for that.
    return GitHubAppKind.COMMIT_SIGNING
}

interface BatchChangesCreateGitHubAppPageProps {
    authenticatedUser: AuthenticatedUser
}

export const BatchChangesCreateGitHubAppPage: FC<BatchChangesCreateGitHubAppPageProps> = ({ authenticatedUser }) => {
    const location = useLocation()
    const searchParams = new URLSearchParams(location.search)
    const baseURL = searchParams.get('baseURL')

    const kind = computeGitHubAppKind(searchParams.get('kind') || GitHubAppKind.COMMIT_SIGNING)
    const isKindCredential = kind !== GitHubAppKind.COMMIT_SIGNING

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
                const existingURL = isKindCredential
                    ? node.externalServiceURL
                    : node.commitSigningConfiguration?.baseURL
                if (!existingURL) {
                    return false
                }

                return new URL(existingURL).hostname === asURL.hostname
            })
            const errorMsg = `A ${
                isKindCredential ? 'GitHub App' : 'commit signing'
            } integration for the code host at this URL already exists.`
            return isDuplicate ? errorMsg : true
        },
        [connection, isKindCredential]
    )
    const pageTitle = isKindCredential
        ? 'Create GitHub App for Batch Changes Credential'
        : 'Create GitHub App for commit signing'
    const defaultAppName = isKindCredential ? 'Batch Changes GitHub App' : 'Sourcegraph Commit Signing'
    return (
        <CreateGitHubAppPage
            defaultEvents={DEFAULT_EVENTS}
            defaultPermissions={DEFAULT_PERMISSIONS}
            pageTitle={pageTitle}
            headerDescription={
                <>
                    Register a GitHub App to enable Sourcegraph {isKindCredential ? 'create' : 'sign commits for'} Batch
                    Change changesets on your behalf.
                    {/* TODO (@BolajiOlajide) update link here for credential github app */}
                    <Link to="/help/admin/config/batch_changes#commit-signing-for-github" className="ml-1">
                        See how GitHub App configuration works.
                    </Link>
                </>
            }
            headerAnnotation={<FeedbackBadge status="beta" feedback={{ mailto: 'support@sourcegraph.com' }} />}
            appDomain={GitHubAppDomain.BATCHES}
            appKind={kind}
            defaultAppName={defaultAppName}
            baseURL={baseURL?.length ? baseURL : undefined}
            validateURL={validateURL}
            telemetryRecorder={noOpTelemetryRecorder}
            authenticatedUser={authenticatedUser}
        />
    )
}
