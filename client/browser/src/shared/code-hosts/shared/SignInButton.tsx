import React, { useCallback } from 'react'

import { interval, type Observable } from 'rxjs'
import { switchMap, filter, take, tap } from 'rxjs/operators'

import { createURLWithUTM } from '@sourcegraph/shared/src/tracking/utm'
import { useEventObservable } from '@sourcegraph/wildcard'

import { SourcegraphIconButton } from '../../components/SourcegraphIconButton'
import { getPlatformName } from '../../util/context'

export const SignInButton: React.FunctionComponent<
    React.PropsWithChildren<{
        className?: string
        iconClassName?: string
        sourcegraphURL: string
        /**
         * Gets called when the user closed the sign in tab.
         * Does not guarantee the sign in was sucessful.
         */
        onSignInClose?: () => void
    }>
> = ({ className, iconClassName, sourcegraphURL, onSignInClose }) => {
    const signInUrl = createURLWithUTM(new URL('/sign-in?close=true', sourcegraphURL), {
        utm_source: getPlatformName(),
        utm_campaign: 'sign-in-button',
    }).href

    const [nextSignInClick] = useEventObservable(
        useCallback(
            (signInClicks: Observable<React.MouseEvent>) =>
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
            [onSignInClose, signInUrl]
        )
    )

    return (
        <SourcegraphIconButton
            href={signInUrl}
            label="Sign in to Sourcegraph"
            title="Sign into Sourcegraph to get hover tooltips, go to definition and more"
            ariaLabel="Sign into Sourcegraph to get hover tooltips, go to definition and more"
            className={className}
            iconClassName={iconClassName}
            onClick={nextSignInClick}
        />
    )
}
