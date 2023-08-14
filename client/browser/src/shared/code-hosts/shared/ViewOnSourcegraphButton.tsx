import React from 'react'

import classNames from 'classnames'
import { snakeCase } from 'lodash'

import { type ErrorLike, isErrorLike } from '@sourcegraph/common'
import { isHTTPAuthError } from '@sourcegraph/http-client'
import { createURLWithUTM } from '@sourcegraph/shared/src/tracking/utm'

import { SourcegraphIconButton, type SourcegraphIconButtonProps } from '../../components/SourcegraphIconButton'
import { getPlatformName, isDefaultSourcegraphUrl } from '../../util/context'

import type { CodeHostContext } from './codeHost'
import { SignInButton } from './SignInButton'

import styles from './ViewOnSourcegraphButton.module.scss'

export interface ViewOnSourcegraphButtonClassProps {
    className?: string
    iconClassName?: string
}

interface ViewOnSourcegraphButtonProps
    extends ViewOnSourcegraphButtonClassProps,
        Pick<ConfigureSourcegraphButtonProps, 'codeHostType' | 'onConfigureSourcegraphClick'> {
    context: CodeHostContext
    sourcegraphURL: string
    userSettingsURL?: string
    minimalUI: boolean
    repoExistsOrError?: boolean | ErrorLike
    showSignInButton?: boolean

    /**
     * A callback for when the user finished a sign in flow.
     * This does not guarantee the sign in was successful.
     */
    onSignInClose?: () => void
}

export const ViewOnSourcegraphButton: React.FunctionComponent<
    React.PropsWithChildren<ViewOnSourcegraphButtonProps>
> = ({
    codeHostType,
    repoExistsOrError,
    sourcegraphURL,
    userSettingsURL,
    context,
    minimalUI,
    onConfigureSourcegraphClick,
    showSignInButton,
    onSignInClose,
    className,
    iconClassName,
}) => {
    className = classNames('open-on-sourcegraph', className)
    const mutedIconClassName = classNames(styles.iconMuted, iconClassName)
    const commonProps: Partial<SourcegraphIconButtonProps> = {
        className,
        iconClassName,
    }

    const { rawRepoName, revision, privateRepository } = context

    // Show nothing while loading
    if (repoExistsOrError === undefined) {
        return null
    }

    const url = createURLWithUTM(new URL(`/${rawRepoName}${revision ? `@${revision}` : ''}`, sourcegraphURL), {
        utm_source: getPlatformName(),
        utm_campaign: 'view-on-sourcegraph',
    }).href

    if (isErrorLike(repoExistsOrError)) {
        // If the problem is the user is not signed in, show a sign in CTA (if not shown elsewhere)
        if (isHTTPAuthError(repoExistsOrError)) {
            if (showSignInButton) {
                return <SignInButton {...commonProps} sourcegraphURL={sourcegraphURL} onSignInClose={onSignInClose} />
            }
            // Sign in button may already be shown elsewhere on the page
            return null
        }

        const commonErrorCaseProps: Partial<SourcegraphIconButtonProps> = {
            ...commonProps,
            // If we are not running in the browser extension where we can open the options menu,
            // open the documentation for how to configure the code host we are on.
            href: new URL(snakeCase(codeHostType), 'https://docs.sourcegraph.com/integration/').href,
            // onClick can call preventDefault() to prevent that and take a different action (opening the options menu).
            onClick: onConfigureSourcegraphClick,
        }

        // If there was an unexpected error, show it in the tooltip.
        // Still link to the Sourcegraph instance in native integrations
        // as that might explain the error (e.g. not reachable).
        // In the browser extension, let the onConfigureSourcegraphClick handler can handle this.
        return (
            <SourcegraphIconButton
                {...commonErrorCaseProps}
                iconClassName={mutedIconClassName}
                href={url}
                label="Error"
                title={repoExistsOrError.message}
                ariaLabel={repoExistsOrError.message}
            />
        )
    }

    // If the repository does not exist, communicate that to explain why e.g. code navigation does not work
    if (!repoExistsOrError) {
        if (isDefaultSourcegraphUrl(sourcegraphURL) && privateRepository && userSettingsURL) {
            return <ConfigureSourcegraphButton {...commonProps} codeHostType={codeHostType} href={userSettingsURL} />
        }

        return (
            <SourcegraphIconButton
                {...commonProps}
                href={url} // Still link to the repository (which will show a not found page, and can link to further actions)
                iconClassName={mutedIconClassName}
                label="Repository not found"
                title={`The repository does not exist on the configured Sourcegraph instance ${sourcegraphURL}`}
                ariaLabel={`The repository does not exist on the configured Sourcegraph instance ${sourcegraphURL}`}
            />
        )
    }

    // Otherwise don't render anything in minimal UI mode
    if (minimalUI) {
        return null
    }

    // Render a "View on Sourcegraph" button
    return (
        <SourcegraphIconButton
            {...commonProps}
            href={url}
            title="View repository on Sourcegraph"
            ariaLabel="View repository on Sourcegraph"
        />
    )
}
interface ConfigureSourcegraphButtonProps extends Partial<SourcegraphIconButtonProps> {
    codeHostType: string
    onConfigureSourcegraphClick?: React.MouseEventHandler<HTMLAnchorElement>
}

export const ConfigureSourcegraphButton: React.FunctionComponent<
    React.PropsWithChildren<ConfigureSourcegraphButtonProps>
> = ({ onConfigureSourcegraphClick, codeHostType, ...commonProps }) => (
    <SourcegraphIconButton
        {...commonProps}
        href={commonProps.href || new URL(snakeCase(codeHostType), 'https://docs.sourcegraph.com/integration/').href}
        onClick={onConfigureSourcegraphClick}
        label="Configure Sourcegraph"
        title="Set up Sourcegraph for search and code navigation on private repositories"
        ariaLabel="Set up Sourcegraph for search and code navigation on private repositories"
    />
)
