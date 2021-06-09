import InfoIcon from 'mdi-react/InfoCircleIcon'
import React, { FunctionComponent, useState, useEffect } from 'react'
import { Link } from 'react-router-dom'

import { DismissibleAlert, isAlertDismissed, dismissAlert } from '../components/DismissibleAlert'

interface Props {
    isUserMissingGitHubPrivateScope: boolean | null
}

export const GITHUB_SCOPE_ALERT_KEY = 'GitHubPrivateScopeAlert'

/**
 * A global alert telling authenticated users if they need re-add GitHub code
 * host connection again to access private repositories
 */
export const GitHubScopeAlert: FunctionComponent<Props> = ({ isUserMissingGitHubPrivateScope }) => {
    const [shouldDisplayAlert, setShouldDisplayAlert] = useState(false)

    useEffect(() => {
        if (isUserMissingGitHubPrivateScope !== null && !isAlertDismissed(GITHUB_SCOPE_ALERT_KEY)) {
            if (isUserMissingGitHubPrivateScope) {
                setShouldDisplayAlert(true)
            } else {
                // if the user has all of the required GitHub scopes -
                // never check for external service scopes and don't show the
                // alert
                dismissAlert(GITHUB_SCOPE_ALERT_KEY)
            }
        }
    }, [isUserMissingGitHubPrivateScope])

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
