import classNames from 'classnames'
import React, { useMemo, useCallback } from 'react'
import { render } from 'react-dom'
import { Observable, interval } from 'rxjs'
import { switchMap, catchError, filter, tap, take } from 'rxjs/operators'
import { SourcegraphIconButton } from '../../shared/components/Button'
import { DEFAULT_SOURCEGRAPH_URL } from '../../shared/util/context'
import { CodeHost, CodeHostContext } from './code_intelligence'
import { asError } from '../../../../shared/src/util/errors'
import { useObservable, useEventObservable } from '../../../../shared/src/util/useObservable'
import { failedWithHTTPStatus } from '../../../../shared/src/backend/fetch'

export interface ViewOnSourcegraphButtonClassProps {
    className?: string
    iconClassName?: string
}

interface ViewOnSourcegraphButtonProps extends ViewOnSourcegraphButtonClassProps {
    context: CodeHostContext
    sourcegraphURL: string
    ensureRepoExists: (context: CodeHostContext, sourcegraphUrl: string) => Observable<boolean>
    onConfigureSourcegraphClick?: () => void
    minimalUI: boolean
}

export const ViewOnSourcegraphButton: React.FunctionComponent<ViewOnSourcegraphButtonProps> = ({
    ensureRepoExists,
    sourcegraphURL,
    context,
    minimalUI,
    onConfigureSourcegraphClick,
    className,
    iconClassName,
}) => {
    /** Whether or not the repo exists on the configured Sourcegraph instance. */
    const repoExistsOrError = useObservable(
        useMemo(() => ensureRepoExists(context, sourcegraphURL).pipe(catchError(error => [asError(error)])), [
            context,
            ensureRepoExists,
            sourcegraphURL,
        ])
    )

    const signInUrl = new URL('/sign-in', sourcegraphURL)
    signInUrl.searchParams.set('close', 'true')
    const [nextSignInClick] = useEventObservable(
        useCallback(
            (events: Observable<React.MouseEvent>) =>
                events.pipe(
                    switchMap(event => {
                        const tab = window.open(signInUrl.href, '_blank')
                        if (!tab) {
                            return []
                        }
                        event.preventDefault()
                        return interval(300).pipe(
                            filter(() => tab.closed),
                            take(1)
                        )
                    }),
                    tap(() => location.reload())
                ),
            [signInUrl]
        )
    )

    if (repoExistsOrError === undefined) {
        return null
    }
    className = classNames('open-on-sourcegraph', className)

    if (failedWithHTTPStatus(repoExistsOrError, 401)) {
        return (
            <SourcegraphIconButton
                url={signInUrl.href}
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

    const url = `${sourcegraphURL}/${context.rawRepoName}${context.rev ? `@${context.rev}` : ''}`
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
