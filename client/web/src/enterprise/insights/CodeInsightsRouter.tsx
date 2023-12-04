import type { FC } from 'react'

import type { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { lazyComponent } from '@sourcegraph/shared/src/util/lazyComponent'

import type { AuthenticatedUser } from '../../auth'

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
    const { authenticatedUser, telemetryService } = props

    if (!window.context?.codeInsightsEnabled) {
        return (
            <CodeInsightsDotComGetStartedLazy
                telemetryService={telemetryService}
                authenticatedUser={authenticatedUser}
            />
        )
    }

    return <CodeInsightsAppLazyRouter authenticatedUser={authenticatedUser} telemetryService={telemetryService} />
}
