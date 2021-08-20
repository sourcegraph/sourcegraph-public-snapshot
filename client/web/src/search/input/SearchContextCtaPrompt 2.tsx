import classNames from 'classnames'
import React from 'react'

import { Link } from '@sourcegraph/shared/src/components/Link'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { AuthenticatedUser } from '../../auth'
import { Badge } from '../../components/Badge'

import styles from './SearchContextCtaPrompt.module.scss'

export interface SearchContextCtaPromptProps extends TelemetryProps {
    authenticatedUser: AuthenticatedUser | null
    hasUserAddedExternalServices: boolean
}

export const SearchContextCtaPrompt: React.FunctionComponent<SearchContextCtaPromptProps> = ({
    authenticatedUser,
    hasUserAddedExternalServices,
    telemetryService,
}) => {
    const repositoriesVisibility =
        window.context.externalServicesUserMode === 'all' ||
        authenticatedUser?.tags.includes('AllowUserExternalServicePrivate')
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
        : `/sign-up?src=Context&returnTo=${encodeURIComponent('/user/settings/repositories')}`

    const onClick = (): void => {
        const authenticatedActionKind = hasUserAddedExternalServices ? 'AddRepositories' : 'ConnectCodeHost'
        const actionKind = authenticatedUser ? authenticatedActionKind : 'SignUp'
        telemetryService.log(`SearchContextCtaPrompt${actionKind}Click`)
    }

    return (
        <div className={styles.searchContextCtaPrompt}>
            <div className={styles.searchContextCtaPromptTitle}>
                <Badge className="mr-1" status="new" />
                <span>Search the code you care about</span>
            </div>
            <div className="text-muted">{copyText}</div>
            <Link
                className={classNames('btn btn-primary', styles.searchContextCtaPromptButton)}
                to={linkTo}
                onClick={onClick}
            >
                {buttonText}
            </Link>
        </div>
    )
}
