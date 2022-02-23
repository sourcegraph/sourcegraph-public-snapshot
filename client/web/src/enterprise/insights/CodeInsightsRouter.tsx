import React from 'react'

import { Settings } from '@sourcegraph/shared/src/schema/settings.schema'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { lazyComponent } from '@sourcegraph/shared/src/util/lazyComponent'

import { AuthenticatedUser } from '../../auth'

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
export interface CodeInsightsRouterProps extends SettingsCascadeProps<Settings>, TelemetryProps {
    /**
     * Authenticated user info, Used to decide where code insight will appear
     * in personal dashboard (private) or in organisation dashboard (public)
     */
    authenticatedUser: AuthenticatedUser
    isSourcegraphDotCom: boolean
}

export const CodeInsightsRouter: React.FunctionComponent<CodeInsightsRouterProps> = props => {
    if (props.isSourcegraphDotCom) {
        return <CodeInsightsDotComGetStartedLazy telemetryService={props.telemetryService} />
    }

    return <CodeInsightsAppLazyRouter {...props} />
}
