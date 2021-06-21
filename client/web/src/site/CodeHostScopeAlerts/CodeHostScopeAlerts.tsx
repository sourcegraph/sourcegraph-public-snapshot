import React, { FunctionComponent } from 'react'
import { Link } from 'react-router-dom'

import { DismissibleAlert } from '../../components/DismissibleAlert'
import { githubRepoScopeRequired, gitlabAPIScopeRequired } from '../../user/settings/cloud-ga'

import { useCodeHostScopeContext } from './CodeHostScopeProvider'

interface Props {
    authenticatedUser: { id: string; tags: string[] } | null
}

export const GITHUB_SCOPE_ALERT_KEY = 'GitHubPrivateScopeAlert'
export const GITLAB_SCOPE_ALERT_KEY = 'GitLabPrivateScopeAlert'

/**
 * A global alert telling authenticated users if they need to update GitHub code
 * host token to access the private repositories.
 */
export const CodeHostScopeAlerts: FunctionComponent<Props> = ({ authenticatedUser }) => {
    const { scopes } = useCodeHostScopeContext()

    if (!authenticatedUser || scopes === null) {
        return null
    }

    if (!githubRepoScopeRequired(authenticatedUser.tags, scopes.github)) {
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

/**
 * A global alert telling authenticated users if they need to update GitLab code
 * host token to access the private repositories.
 */
export const GitLabScopeAlert: FunctionComponent<Props> = ({ authenticatedUser }) => {
    const { scopes } = useCodeHostScopeContext()

    if (!authenticatedUser || scopes === null) {
        return null
    }

    if (!gitlabAPIScopeRequired(authenticatedUser.tags, scopes.gitlab)) {
        return null
    }

    return (
        <DismissibleAlert partialStorageKey={GITLAB_SCOPE_ALERT_KEY} className="alert-info global-alerts__alert">
            <span>
                Update your <Link to="/user/settings/code-hosts">GitLab code host connection</Link> to search private
                code with Sourcegraph.
            </span>
        </DismissibleAlert>
    )
}
