import React from 'react'

import classNames from 'classnames'

import { AuthenticatedUser } from '@sourcegraph/shared/src/auth'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { buildGetStartedURL } from '@sourcegraph/shared/src/util/url'
import { Button, ProductStatusBadge, Link, ButtonLink } from '@sourcegraph/wildcard'

import styles from './SearchContextCtaPrompt.module.scss'

export interface SearchContextCtaPromptProps extends TelemetryProps {
    authenticatedUser: AuthenticatedUser | null
    hasUserAddedExternalServices: boolean
    onDismiss: () => void
    /** Set in JSContext so only available to the web app. */
    isExternalServicesUserModeAll?: boolean
}

export const SearchContextCtaPrompt: React.FunctionComponent<React.PropsWithChildren<SearchContextCtaPromptProps>> = ({
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

    const buttonText = hasUserAddedExternalServices ? 'Add repositories' : 'Connect with code host'

    const linkTo = authenticatedUser
        ? hasUserAddedExternalServices
            ? `/users/${authenticatedUser.username}/settings/repositories`
            : `/users/${authenticatedUser.username}/settings/code-hosts`
        : null

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

            {authenticatedUser && linkTo !== null ? (
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
            ) : (
                <ButtonLink
                    className={styles.searchContextCtaPromptButton}
                    to={buildGetStartedURL('search-context-cta', '/user/settings/repositories')}
                    onClick={onClick}
                    variant="primary"
                    size="sm"
                >
                    Get started
                </ButtonLink>
            )}
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
