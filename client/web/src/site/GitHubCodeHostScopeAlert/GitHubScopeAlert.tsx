import InfoIcon from 'mdi-react/InfoCircleIcon'
import React, { FunctionComponent } from 'react'
import { Link } from 'react-router-dom'

import { DismissibleAlert } from '../../components/DismissibleAlert'
import { githubRepoScopeRequired } from '../../user/settings/cloud-ga'

import { useGitHubScopeContext } from './GithubScopeProvider'

interface Props {
    authenticatedUser: { id: string; tags: string[] } | null
}

export const GITHUB_SCOPE_ALERT_KEY = 'GitHubPrivateScopeAlert'

/**
 * A global alert telling authenticated users if they need to update GitHub code
 * host token to access the private repositories.
 */
export const GitHubScopeAlert: FunctionComponent<Props> = ({ authenticatedUser }) => {
    const { scopes } = useGitHubScopeContext()

    const shouldTryToDisplayAlert = (): boolean => {
        if (!authenticatedUser || scopes === null) {
            return false
        }

        return githubRepoScopeRequired(authenticatedUser.tags, scopes)
    }

    return shouldTryToDisplayAlert() ? (
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
