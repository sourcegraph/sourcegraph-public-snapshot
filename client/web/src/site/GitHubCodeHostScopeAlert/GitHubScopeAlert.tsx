import InfoIcon from 'mdi-react/InfoCircleIcon'
import React, { FunctionComponent, useState, useEffect, useCallback, useContext } from 'react'
import { Link } from 'react-router-dom'

import { DismissibleAlert, isAlertDismissed, dismissAlert } from '../../components/DismissibleAlert'
import { githubRepoScopeRequired } from '../../user/settings/cloud-ga'

import { GitHubScopeContext } from './GithubScopeContext'

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
    const scopes = useContext(GitHubScopeContext)

    const checkGitHubServiceScope = useCallback(() => {
        if (authenticatedUser && !isAlertDismissed(GITHUB_SCOPE_ALERT_KEY)) {
            // check whether we need to prompt the user to update their scope
            if (scopes) {
                const hasMissingScope = githubRepoScopeRequired(authenticatedUser.tags, scopes)
                if (hasMissingScope) {
                    setShouldDisplayAlert(true)
                } else {
                    // if the user has required GitHub scopes - never check for
                    // external service scopes and don't show the alert
                    dismissAlert(GITHUB_SCOPE_ALERT_KEY)
                }
            }
        }
    }, [authenticatedUser, scopes])

    useEffect(() => {
        checkGitHubServiceScope()
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
