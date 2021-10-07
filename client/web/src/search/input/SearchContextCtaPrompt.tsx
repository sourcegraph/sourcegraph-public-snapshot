import classNames from 'classnames'
import React, { useState } from 'react'

import { Link } from '@sourcegraph/shared/src/components/Link'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { AuthenticatedUser } from '../../auth'
import { Badge } from '../../components/Badge'

import styles from './SearchContextCtaPrompt.module.scss'

export interface SearchContextCtaPromptProps extends TelemetryProps {
    authenticatedUser: AuthenticatedUser | null
    hasUserAddedExternalServices: boolean
    onDismiss: (permanent: boolean) => void
}

export const SearchContextCtaPrompt: React.FunctionComponent<SearchContextCtaPromptProps> = ({
    authenticatedUser,
    hasUserAddedExternalServices,
    telemetryService,
    onDismiss,
}) => {
    const [ctaPermanentlyDismissed, setCtaPermanentlyDismissed] = useState(false)

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

    const onDismissClick = (): void => {
        telemetryService.log(
            'SearchContextCtaPromptDismissClick',
            { permanent: ctaPermanentlyDismissed },
            { permanent: ctaPermanentlyDismissed }
        )
        onDismiss(ctaPermanentlyDismissed)
    }

    return (
        <div className={styles.searchContextCtaPrompt}>
            <div className={styles.searchContextCtaPromptTitle}>
                <Badge className="mr-1" status="new" />
                <span>Search the code you care about</span>
            </div>
            <div className="text-muted">{copyText}</div>

            <label className="d-flex align-items-center mt-2">
                <input
                    type="checkbox"
                    className="mr-2"
                    checked={ctaPermanentlyDismissed}
                    onChange={event => setCtaPermanentlyDismissed(event.target.checked)}
                />
                Don't show this again
            </label>

            <Link
                className={classNames('btn btn-primary', styles.searchContextCtaPromptButton)}
                to={linkTo}
                onClick={onClick}
            >
                {buttonText}
            </Link>
            <button
                type="button"
                className={classNames('btn btn-secondary ml-2', styles.searchContextCtaPromptButton)}
                onClick={onDismissClick}
            >
                Maybe later
            </button>
        </div>
    )
}
