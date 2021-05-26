import React from 'react'

import { Link } from '@sourcegraph/shared/src/components/Link'

import { AuthenticatedUser } from '../../auth'
import { Badge } from '../../components/Badge'

import styles from './SearchContextCtaPrompt.module.scss'

export interface SearchContextCtaPromptProps {
    authenticatedUser: AuthenticatedUser | null
    hasUserAddedExternalServices: boolean
}

export const SearchContextCtaPrompt: React.FunctionComponent<SearchContextCtaPromptProps> = ({
    authenticatedUser,
    hasUserAddedExternalServices,
}) => {
    const repositoriesVisibility = authenticatedUser?.tags.includes('AllowUserExternalServicePrivate')
        ? 'repositories'
        : 'public repositories'

    const copyText = authenticatedUser
        ? `Add your ${repositoriesVisibility} from GitHub or Gitlab to Sourcegraph and power up your searches with your personal search context.`
        : `Connect with GitHub or Gitlab to add your ${repositoriesVisibility} to Sourcegraph and power up your searches with your personal search context.`

    const buttonText = authenticatedUser
        ? hasUserAddedExternalServices
            ? 'Add repositories'
            : 'Connect with code host'
        : 'Sign up for Sourcegraph'
    const linkTo = authenticatedUser
        ? hasUserAddedExternalServices
            ? `/users/${authenticatedUser.username}/settings/repositories`
            : `/users/${authenticatedUser.username}/settings/code-hosts`
        : '/sign-up'

    return (
        <div className={styles.searchContextCtaPrompt}>
            <div className={styles.searchContextCtaPromptTitle}>
                <Badge className="mr-1" status="new" />
                <span>Search the code you care about</span>
            </div>
            <div className="text-muted">{copyText}</div>
            <div className={styles.searchContextCtaPromptButton}>
                <Link className="btn btn-primary" to={linkTo}>
                    {buttonText}
                </Link>
            </div>
        </div>
    )
}
