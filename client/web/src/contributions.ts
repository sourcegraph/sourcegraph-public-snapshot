import { useEffect, useRef, useState } from 'react'

import { useLocation, useNavigate } from 'react-router-dom-v5-compat'
import { Subscription } from 'rxjs'

import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { registerHoverContributions } from '@sourcegraph/shared/src/hover/actions'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'

interface Props extends ExtensionsControllerProps, PlatformContextProps {}

/**
 * A component that registers global contributions. It is implemented as a React component so that its
 * registrations use the React lifecycle.
 */
export function GlobalContributions(props: Props): null {
    const { extensionsController, platformContext } = props

    const location = useLocation()
    const navigate = useNavigate()

    // Location may be used by the hover contributions after they are
    // initialized. To avoid stale location, we need to keep it in a ref.
    const locationRef = useRef(location)
    useEffect(() => {
        locationRef.current = location
    }, [location])

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
                    historyOrNavigate: navigate,
                    getLocation: () => locationRef.current,
                    extensionsController,
                    locationAssign: globalThis.location.assign.bind(globalThis.location),
                })
            )
        }
        return () => subscriptions.unsubscribe()
    }, [extensionsController, navigate, platformContext])

    // Throw error to the <ErrorBoundary />
    if (error) {
        throw error
    }

    return null
}
