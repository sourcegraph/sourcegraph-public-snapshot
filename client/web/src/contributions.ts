import { useEffect, useRef, useState } from 'react'

import { type NavigateFunction, useLocation, useNavigate } from 'react-router-dom'
import { Subscription } from 'rxjs'

import type { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { registerHoverContributions } from '@sourcegraph/shared/src/hover/actions'
import type { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { TelemetryV2Props } from '@sourcegraph/shared/src/telemetry'

interface Props extends ExtensionsControllerProps, PlatformContextProps, TelemetryV2Props {}

/**
 * A component that registers global contributions. It is implemented as a React component so that its
 * registrations use the React lifecycle.
 */
export function GlobalContributions(props: Props): null {
    const { extensionsController, platformContext } = props

    const location = useLocation()
    const navigate = useNavigate()

    // Location and navigate may be used by the hover contributions after they
    // are initialized and closed over. To avoid stale data, we keep them in
    // refs.
    const locationRef = useRef(location)
    const navigateRef = useRef(navigate)
    useEffect(() => {
        locationRef.current = location
        navigateRef.current = navigate
    }, [location, navigate])

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
            const historyOrNavigate: NavigateFunction = ((to: any, options: any): void =>
                navigateRef.current?.(to, options)) as any
            subscriptions.add(
                registerHoverContributions({
                    platformContext,
                    historyOrNavigate,
                    getLocation: () => locationRef.current,
                    extensionsController,
                    locationAssign: globalThis.location.assign.bind(globalThis.location),
                })
            )
        }
        return () => subscriptions.unsubscribe()
    }, [extensionsController, platformContext])

    // Throw error to the <ErrorBoundary />
    if (error) {
        throw error
    }

    return null
}
