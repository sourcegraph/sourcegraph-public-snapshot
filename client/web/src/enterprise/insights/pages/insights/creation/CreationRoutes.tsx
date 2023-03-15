import { FC } from 'react'

import { Routes, Route } from 'react-router-dom'

import { useExperimentalFeatures } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { lazyComponent } from '@sourcegraph/shared/src/util/lazyComponent'

import { InsightCreationPageType } from './InsightCreationPage'

const IntroCreationLazyPage = lazyComponent(() => import('./intro/IntroCreationPage'), 'IntroCreationPage')
const InsightCreationLazyPage = lazyComponent(() => import('./InsightCreationPage'), 'InsightCreationPage')

interface CreationRoutesProps extends TelemetryProps {
    isSourcegraphApp: boolean
}

/**
 * Code insight sub-router for the creation area/routes.
 * Renders code insights creation routes (insight creation UI pages, creation intro page)
 */
export const CreationRoutes: FC<CreationRoutesProps> = props => {
    const { telemetryService, isSourcegraphApp } = props

    const codeInsightsCompute = useExperimentalFeatures(settings => settings.codeInsightsCompute)

    return (
        <Routes>
            <Route
                index={true}
                element={
                    <IntroCreationLazyPage telemetryService={telemetryService} isSourcegraphApp={isSourcegraphApp} />
                }
            />

            <Route
                path="search"
                element={
                    <InsightCreationLazyPage
                        mode={InsightCreationPageType.Search}
                        telemetryService={telemetryService}
                        isSourcegraphApp={isSourcegraphApp}
                    />
                }
            />

            <Route
                path="capture-group"
                element={
                    <InsightCreationLazyPage
                        mode={InsightCreationPageType.CaptureGroup}
                        telemetryService={telemetryService}
                        isSourcegraphApp={isSourcegraphApp}
                    />
                }
            />

            <Route
                path="lang-stats"
                element={
                    <InsightCreationLazyPage
                        mode={InsightCreationPageType.LangStats}
                        telemetryService={telemetryService}
                        isSourcegraphApp={isSourcegraphApp}
                    />
                }
            />

            {codeInsightsCompute && (
                <Route
                    path="group-results"
                    element={
                        <InsightCreationLazyPage
                            mode={InsightCreationPageType.Compute}
                            telemetryService={telemetryService}
                            isSourcegraphApp={isSourcegraphApp}
                        />
                    }
                />
            )}
        </Routes>
    )
}
