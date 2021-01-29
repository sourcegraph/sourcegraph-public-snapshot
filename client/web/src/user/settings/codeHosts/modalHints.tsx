import React from 'react'
import { Link } from '../../../../../shared/src/components/Link'
import { ExternalServiceKind } from '../../../graphql-operations'

export const hints: Partial<Record<ExternalServiceKind, React.ReactFragment>> = {
    [ExternalServiceKind.GITHUB]: (
        <small>
            <Link
                to="https://docs.github.com/en/free-pro-team@latest/github/authenticating-to-github/creating-a-personal-access-token"
                target="_blank"
                rel="noopener noreferrer"
            >
                Create a new access token
            </Link>
            <span className="text-muted"> on GitHub.com with repo or public_repo scope.</span>
        </small>
    ),
    [ExternalServiceKind.GITLAB]: (
        <small>
            <Link
                to="https://docs.gitlab.com/ee/user/profile/personal_access_tokens.html#creating-a-personal-access-token"
                target="_blank"
                rel="noopener noreferrer"
            >
                Create a new access token
            </Link>
            <span className="text-muted"> on GitLab.com with read_user, read_api, and read_repository scope.</span>
        </small>
    ),
}
