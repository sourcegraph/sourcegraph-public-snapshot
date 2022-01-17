import classNames from 'classnames'
import React from 'react'

import { AuthenticatedUser } from '@sourcegraph/shared/src/auth'
import { Link } from '@sourcegraph/shared/src/components/Link'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ProductStatusBadge, Button } from '@sourcegraph/wildcard'

import styles from './SearchContextCtaPrompt.module.scss'

export interface SearchContextCtaPromptProps extends TelemetryProps {
    authenticatedUser: AuthenticatedUser | null
    hasUserAddedExternalServices: boolean
    onDismiss: () => void
    /** Set in JSContext only available to the web app. */
    isExternalServicesUserModeAll?: boolean
}

export const SearchContextCtaPrompt: React.FunctionComponent<SearchContextCtaPromptProps> = ({
    authenticatedUser,
    hasUserAddedExternalServices,
    telemetryService,
    isExternalServicesUserModeAll,
    onDismiss,
}) => {
    const repositoriesVisibility =
        isExternalServicesUserModeAll || authenticatedUser?.tags.includes('AllowUserExternalServicePrivate')
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

    const onDismissClick = (): void => {
        telemetryService.log('SearchContextCtaPromptDismissClick')
        onDismiss()
    }

    return (
        <div className={classNames(styles.searchContextCtaPrompt)}>
            <div className={styles.searchContextCtaPromptTitle}>
                <ProductStatusBadge className="mr-1" status="new" />
                <span>Search the code you care about</span>
            </div>
            <div className="text-muted">{copyText}</div>

            <Button
                className={styles.searchContextCtaPromptButton}
                to={linkTo}
                onClick={onClick}
                variant="primary"
                size="sm"
                as={Link}
            >
                {buttonText}
            </Button>
            <Button
                className={classNames('border-0 ml-2', styles.searchContextCtaPromptButton)}
                onClick={onDismissClick}
                outline={true}
                variant="secondary"
                size="sm"
            >
                Don't show this again
            </Button>
        </div>
    )
}
