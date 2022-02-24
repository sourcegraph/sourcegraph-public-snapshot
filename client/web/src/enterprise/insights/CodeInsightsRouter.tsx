import React from 'react'

import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { Settings } from '@sourcegraph/shared/src/schema/settings.schema'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { lazyComponent } from '@sourcegraph/shared/src/util/lazyComponent'

import { AuthenticatedUser } from '../../auth'
import { withAuthenticatedUser } from '../../auth/withAuthenticatedUser'

const CodeInsightsAppLazyRouter = lazyComponent(() => import('./CodeInsightsAppRouter'), 'CodeInsightsAppRouter')

const CodeInsightsDotComGetStartedLazy = lazyComponent(
    () => import('./pages/dot-com-get-started/CodeInsightsDotComGetStarted'),
    'CodeInsightsDotComGetStarted'
)

/**
 * This interface has to receive union type props derived from all child components
 * Because we need to pass all required prop from main Sourcegraph.tsx component to
 * subcomponents withing app tree.
 */
export interface CodeInsightsRouterProps extends SettingsCascadeProps<Settings>, PlatformContextProps, TelemetryProps {
    /**
     * Authenticated user info, Used to decide where code insight will appears
     * in personal dashboard (private) or in organisation dashboard (public)
     */
    authenticatedUser: AuthenticatedUser
    isSourcegraphDotCom: boolean
}

/**
 * Turn on/off the cloud landing page layout. Make sure it's off until GA release will happen.
 */
const CLOUD_LANDING_PAGE = false

/**
 * Main Insight routing component. Main entry point to code insights UI.
 */
export const CodeInsightsRouter = withAuthenticatedUser<CodeInsightsRouterProps>(props => {
    if (props.isSourcegraphDotCom && CLOUD_LANDING_PAGE) {
        return <CodeInsightsDotComGetStartedLazy telemetryService={props.telemetryService} />
    }

    return <CodeInsightsAppLazyRouter {...props} />
})
