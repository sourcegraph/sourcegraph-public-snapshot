import classNames from 'classnames'
import React, { useMemo } from 'react'
import { render } from 'react-dom'
import { Observable, interval, Subject, concat } from 'rxjs'
import { switchMap, catchError, filter, tap, take } from 'rxjs/operators'
import { SourcegraphIconButton } from '../../shared/components/Button'
import { DEFAULT_SOURCEGRAPH_URL } from '../../shared/util/context'
import { CodeHost, CodeHostContext } from './code_intelligence'
import { asError } from '../../../../shared/src/util/errors'
import { useObservable } from '../../../../shared/src/util/useObservable'
import { failedWithHTTPStatus } from '../../../../shared/src/backend/fetch'

export interface ViewOnSourcegraphButtonClassProps {
    className?: string
    iconClassName?: string
}

interface ViewOnSourcegraphButtonProps extends ViewOnSourcegraphButtonClassProps {
    context: CodeHostContext
    sourcegraphURL: string
    minimalUI: boolean
    ensureRepoExists: (context: CodeHostContext, sourcegraphUrl: string) => Observable<boolean>
    onConfigureSourcegraphClick?: () => void

    /**
     * A callback for when the user finished a sign in flow.
     * This does not guarantee the sign in was successful.
     */
    onSignInClose?: () => void
}

export const ViewOnSourcegraphButton: React.FunctionComponent<ViewOnSourcegraphButtonProps> = ({
    ensureRepoExists,
    sourcegraphURL,
    context,
    minimalUI,
    onConfigureSourcegraphClick,
    onSignInClose,
    className,
    iconClassName,
}) => {
    const signInUrl = new URL('/sign-in?close=true', sourcegraphURL).href

    /** Clicks on the "Sign in" CTA (if rendered) */
    const signInClicks = useMemo(() => new Subject<React.MouseEvent>(), [])
    const nextSignInClick = useMemo(() => signInClicks.next.bind(signInClicks), [signInClicks])

    /**
     * Emits when the user closed the sign in tab again.
     * Does not guarantee the sign in was sucessful.
     */
    const signInCloses = useMemo(
        () =>
            signInClicks.pipe(
                switchMap(event => {
                    const tab = window.open(signInUrl, '_blank')
                    if (!tab) {
                        return []
                    }
                    event.preventDefault()
                    return interval(300).pipe(
                        filter(() => tab.closed),
                        take(1)
                    )
                }),
                tap(onSignInClose)
            ),
        [onSignInClose, signInClicks, signInUrl]
    )

    /** Whether or not the repo exists on the configured Sourcegraph instance. */
    const repoExistsOrError = useObservable(
        useMemo(
            () =>
                concat([null], signInCloses).pipe(
                    switchMap(() =>
                        ensureRepoExists(context, sourcegraphURL).pipe(catchError(error => [asError(error)]))
                    )
                ),
            [context, ensureRepoExists, signInCloses, sourcegraphURL]
        )
    )

    if (repoExistsOrError === undefined) {
        return null
    }
    className = classNames('open-on-sourcegraph', className)

    if (failedWithHTTPStatus(repoExistsOrError, 401)) {
        return (
            <SourcegraphIconButton
                url={signInUrl}
                label="Sign in to Sourcegraph"
                ariaLabel="Sign into Sourcegraph to get hover tooltips, go to definition and more"
                className={className}
                iconClassName={iconClassName}
                onClick={nextSignInClick}
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
                ariaLabel="Install Sourcegraph for search and code intelligence on private instance"
                className={className}
                iconClassName={classNames('open-on-sourcegraph__icon--muted', iconClassName)}
                onClick={onConfigureSourcegraphClick}
            />
        )
    }

    const url = new URL(`/${context.rawRepoName}${context.rev ? `@${context.rev}` : ''}`, sourcegraphURL).href
    return (
        <SourcegraphIconButton
            url={url}
            ariaLabel="View repository on Sourcegraph"
            className={className}
            iconClassName={iconClassName}
        />
    )
}

export const renderViewContextOnSourcegraph = ({
    sourcegraphURL,
    getContext,
    ensureRepoExists,
    viewOnSourcegraphButtonClassProps,
    onConfigureSourcegraphClick,
    minimalUI,
}: Pick<
    ViewOnSourcegraphButtonProps,
    'sourcegraphURL' | 'ensureRepoExists' | 'onConfigureSourcegraphClick' | 'minimalUI'
> &
    Required<Pick<CodeHost, 'getContext'>> &
    Pick<CodeHost, 'viewOnSourcegraphButtonClassProps'>) => (mount: HTMLElement): void => {
    render(
        <ViewOnSourcegraphButton
            {...viewOnSourcegraphButtonClassProps}
            context={getContext()}
            minimalUI={minimalUI}
            sourcegraphURL={sourcegraphURL}
            ensureRepoExists={ensureRepoExists}
            onConfigureSourcegraphClick={onConfigureSourcegraphClick}
        />,
        mount
    )
}
