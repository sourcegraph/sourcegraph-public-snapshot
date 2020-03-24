import classNames from 'classnames'
import React from 'react'
import { SourcegraphIconButton } from '../../shared/components/Button'
import { DEFAULT_SOURCEGRAPH_URL } from '../../shared/util/context'
import { CodeHostContext } from './code_intelligence'
import { ErrorLike, isErrorLike } from '../../../../shared/src/util/errors'
import { failedWithHTTPStatus } from '../../../../shared/src/backend/fetch'
import { SignInButton } from './SignInButton'

export interface ViewOnSourcegraphButtonClassProps {
    className?: string
    iconClassName?: string
}

interface ViewOnSourcegraphButtonProps extends ViewOnSourcegraphButtonClassProps {
    getContext: () => CodeHostContext
    sourcegraphURL: string
    minimalUI: boolean
    repoExistsOrError?: boolean | ErrorLike
    showSignInButton?: boolean
    onConfigureSourcegraphClick?: () => void

    /**
     * A callback for when the user finished a sign in flow.
     * This does not guarantee the sign in was successful.
     */
    onSignInClose?: () => void
}

export const ViewOnSourcegraphButton: React.FunctionComponent<ViewOnSourcegraphButtonProps> = ({
    repoExistsOrError,
    sourcegraphURL,
    getContext,
    minimalUI,
    onConfigureSourcegraphClick,
    showSignInButton,
    onSignInClose,
    className,
    iconClassName,
}) => {
    className = classNames('open-on-sourcegraph', className)

    if (repoExistsOrError === undefined) {
        return null
    }
    if (isErrorLike(repoExistsOrError)) {
        if (failedWithHTTPStatus(repoExistsOrError, 401)) {
            if (showSignInButton) {
                return (
                    <SignInButton
                        sourcegraphURL={sourcegraphURL}
                        onSignInClose={onSignInClose}
                        className={className}
                        iconClassName={iconClassName}
                    />
                )
            }
            // Sign in button may already be shown elsewhere on the page
            return null
        }
        return (
            <SourcegraphIconButton
                label="Error"
                title={repoExistsOrError.message}
                ariaLabel={repoExistsOrError.message}
                className={className}
                iconClassName={classNames('open-on-sourcegraph__icon--muted', iconClassName)}
                onClick={onConfigureSourcegraphClick}
            />
        )
    }
    // In minimal UI mode, only show the button as a CTA to sign in
    if (minimalUI) {
        return null
    }

    // If repo doesn't exist and the instance is sourcegraph.com, prompt
    // user to configure Sourcegraph.
    if (!repoExistsOrError && sourcegraphURL === DEFAULT_SOURCEGRAPH_URL && onConfigureSourcegraphClick) {
        return (
            <SourcegraphIconButton
                label="Configure Sourcegraph"
                title="Install Sourcegraph for search and code intelligence on private instance"
                ariaLabel="Install Sourcegraph for search and code intelligence on private instance"
                className={className}
                iconClassName={classNames('open-on-sourcegraph__icon--muted', iconClassName)}
                onClick={onConfigureSourcegraphClick}
            />
        )
    }

    const { rawRepoName, rev } = getContext()
    const url = new URL(`/${rawRepoName}${rev ? `@${rev}` : ''}`, sourcegraphURL).href
    return (
        <SourcegraphIconButton
            url={url}
            title="View repository on Sourcegraph"
            ariaLabel="View repository on Sourcegraph"
            className={className}
            iconClassName={iconClassName}
        />
    )
}
