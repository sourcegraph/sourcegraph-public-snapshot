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

    if (!authenticatedUser || scopes === null) {
        return null
    }

    if (!githubRepoScopeRequired(authenticatedUser.tags, scopes)) {
        return null
    }

    return (
        <DismissibleAlert partialStorageKey={GITHUB_SCOPE_ALERT_KEY} className="alert-info global-alerts__alert">
            <span>
                Update your <Link to="/user/settings/code-hosts">GitHub code host connection</Link> to search private
                code with Sourcegraph.
            </span>
        </DismissibleAlert>
    )
}
