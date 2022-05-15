import React from 'react'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { lazyComponent } from '@sourcegraph/shared/src/util/lazyComponent'

import { AuthenticatedUser } from '../../auth'

import { CodeInsightsBackendContext } from './core'
import { useApi } from './hooks/use-api'

const CodeInsightsAppLazyRouter = lazyComponent(() => import('./CodeInsightsAppRouter'), 'CodeInsightsAppRouter')

const CodeInsightsDotComGetStartedLazy = lazyComponent(
    () => import('./pages/landing/dot-com-get-started/CodeInsightsDotComGetStarted'),
    'CodeInsightsDotComGetStarted'
)

/**
 * This interface has to receive union type props derived from all child components
 * Because we need to pass all required prop from main Sourcegraph.tsx component to
 * subcomponents withing app tree.
 */
export interface CodeInsightsRouterProps extends TelemetryProps {
    /**
     * Authenticated user info, Used to decide where code insight will appear
     * in personal dashboard (private) or in organisation dashboard (public)
     */
    authenticatedUser: AuthenticatedUser | null
    isSourcegraphDotCom: boolean
}

export const CodeInsightsRouter: React.FunctionComponent<React.PropsWithChildren<CodeInsightsRouterProps>> = props => {
    const api = useApi()

    if (!api) {
        return null
    }

    return (
        <CodeInsightsBackendContext.Provider value={api}>
            {props.isSourcegraphDotCom ? (
                <CodeInsightsDotComGetStartedLazy telemetryService={props.telemetryService} />
            ) : (
                <CodeInsightsAppLazyRouter {...props} />
            )}
        </CodeInsightsBackendContext.Provider>
    )
}
