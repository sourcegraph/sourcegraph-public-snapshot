import { FC } from 'react'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { lazyComponent } from '@sourcegraph/shared/src/util/lazyComponent'

import { AuthenticatedUser } from '../../auth'

const CodeInsightsAppLazyRouter = lazyComponent(() => import('./CodeInsightsAppRouter'), 'CodeInsightsAppRouter')

const CodeInsightsDotComGetStartedLazy = lazyComponent(
    () => import('./pages/landing/dot-com-get-started/CodeInsightsDotComGetStarted'),
    'CodeInsightsDotComGetStarted'
)

export interface CodeInsightsRouterProps extends TelemetryProps {
    authenticatedUser: AuthenticatedUser | null
    isSourcegraphDotCom: boolean
}

export const CodeInsightsRouter: FC<CodeInsightsRouterProps> = props => {
    const { authenticatedUser, isSourcegraphDotCom, telemetryService } = props

    if (isSourcegraphDotCom) {
        return <CodeInsightsDotComGetStartedLazy telemetryService={telemetryService} />
    }

    return <CodeInsightsAppLazyRouter authenticatedUser={authenticatedUser} telemetryService={telemetryService} />
}
