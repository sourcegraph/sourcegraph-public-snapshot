import { FC } from 'react'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { lazyComponent } from '@sourcegraph/shared/src/util/lazyComponent'

import { AuthenticatedUser } from '../../auth'

const CodeInsightsAppLazyRouter = lazyComponent(() => import('./CodeInsightsAppRouter'), 'CodeInsightsAppRouter')

export interface CodeInsightsRouterProps extends TelemetryProps {
    authenticatedUser: AuthenticatedUser | null
    isSourcegraphApp: boolean
}

export const CodeInsightsRouter: FC<CodeInsightsRouterProps> = props => {
    const { authenticatedUser, telemetryService, isSourcegraphApp } = props

    return (
        <CodeInsightsAppLazyRouter
            authenticatedUser={authenticatedUser}
            telemetryService={telemetryService}
            isSourcegraphApp={isSourcegraphApp}
        />
    )
}
