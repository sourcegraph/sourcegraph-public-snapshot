import InfoIcon from 'mdi-react/InfoCircleIcon'
import React, { FunctionComponent, useState, useEffect, useCallback } from 'react'
import { Link } from 'react-router-dom'

import { DismissibleAlert, isAlertDismissed, dismissAlert } from '../components/DismissibleAlert'
import { queryExternalServicesScope } from '../components/externalServices/backend'
import { ExternalServiceKind } from '../graphql-operations'
import { githubRepoScopeRequired } from '../user/settings/cloud-ga'

interface Props {
    authenticatedUser: { id: string; tags: string[] } | null
}

export const GITHUB_SCOPE_ALERT_KEY = 'GitHubPrivateScopeAlert'

/**
 * A global alert telling authenticated users if they need re-add GitHub code
 * host connection again to access private repositories
 */
export const GitHubScopeAlert: FunctionComponent<Props> = ({ authenticatedUser }) => {
    const [shouldDisplayAlert, setShouldDisplayAlert] = useState(false)

    const checkGitHubServiceScope = useCallback(async (): Promise<void> => {
        if (authenticatedUser && !isAlertDismissed(GITHUB_SCOPE_ALERT_KEY)) {
            // fetch all code hosts for given user
            const { nodes: fetchedServices } = await queryExternalServicesScope({
                namespace: authenticatedUser.id,
            }).toPromise()

            // check if user has a GitHub code host
            const gitHubService = fetchedServices.find(({ kind }) => kind === ExternalServiceKind.GITHUB)

            // check whether we need to prompt the user to update their scope
            if (gitHubService) {
                const hasMissingScope = githubRepoScopeRequired(authenticatedUser.tags, gitHubService.grantedScopes)
                if (hasMissingScope) {
                    setShouldDisplayAlert(true)
                } else {
                    // if the user has required GitHub scopes - never check for
                    // external service scopes and don't show the alert
                    dismissAlert(GITHUB_SCOPE_ALERT_KEY)
                }
            }
        }
    }, [authenticatedUser])

    useEffect(() => {
        checkGitHubServiceScope().catch(error => console.log(error))
    }, [checkGitHubServiceScope, authenticatedUser])

    return shouldDisplayAlert ? (
        <DismissibleAlert
            partialStorageKey={GITHUB_SCOPE_ALERT_KEY}
            className="alert alert-info d-flex align-items-center"
        >
            <InfoIcon className="redesign-d-none icon-inline mr-2 flex-shrink-0" />
            Update your&nbsp;
            <Link className="site-alert__link" to="/user/settings/code-hosts">
                <span className="underline">GitHub code host connection</span>
            </Link>
            &nbsp;to search private code with Sourcegraph.
        </DismissibleAlert>
    ) : null
}
