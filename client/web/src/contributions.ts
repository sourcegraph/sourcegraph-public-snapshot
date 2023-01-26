import { useEffect, useState } from 'react'

import * as H from 'history'
import { NavigateFunction } from 'react-router-dom-v5-compat'
import { Subscription } from 'rxjs'

import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { registerHoverContributions } from '@sourcegraph/shared/src/hover/actions'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'

interface Props extends ExtensionsControllerProps, PlatformContextProps {
    historyOrNavigate: H.History | NavigateFunction
    location: H.Location
}

/**
 * A component that registers global contributions. It is implemented as a React component so that its
 * registrations use the React lifecycle.
 */
export function GlobalContributions(props: Props): null {
    const { extensionsController, platformContext, historyOrNavigate, location } = props

    const [error, setError] = useState<null | Error>(null)

    useEffect(() => {
        // Lazy-load `highlight/contributions.ts` to make main application bundle ~25kb Gzip smaller.
        import('@sourcegraph/common/src/util/markdown/contributions')
            .then(({ registerHighlightContributions }) => registerHighlightContributions()) // no way to unregister these
            .catch(setError)
    }, [])

    useEffect(() => {
        const subscriptions = new Subscription()
        if (extensionsController !== null) {
            subscriptions.add(
                registerHoverContributions({
                    platformContext,
                    historyOrNavigate,
                    location,
                    extensionsController,
                    locationAssign: globalThis.location.assign.bind(location),
                })
            )
        }
        return () => subscriptions.unsubscribe()
    })

    // Throw error to the <ErrorBoundary />
    if (error) {
        throw error
    }

    return null
}
