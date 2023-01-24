import React from 'react'

import { Switch, useRouteMatch } from 'react-router'
import { CompatRoute } from 'react-router-dom-v5-compat'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { lazyComponent } from '@sourcegraph/shared/src/util/lazyComponent'

import { useExperimentalFeatures } from '../../../../../stores'

import { InsightCreationPageType } from './InsightCreationPage'

const IntroCreationLazyPage = lazyComponent(() => import('./intro/IntroCreationPage'), 'IntroCreationPage')
const InsightCreationLazyPage = lazyComponent(() => import('./InsightCreationPage'), 'InsightCreationPage')

interface CreationRoutesProps extends TelemetryProps {}

/**
 * Code insight sub-router for the creation area/routes.
 * Renders code insights creation routes (insight creation UI pages, creation intro page)
 */
export const CreationRoutes: React.FunctionComponent<React.PropsWithChildren<CreationRoutesProps>> = props => {
    const { telemetryService } = props

    const match = useRouteMatch()
    const { codeInsightsCompute } = useExperimentalFeatures()

    return (
        <Switch>
            <CompatRoute
                exact={true}
                path={`${match.url}`}
                render={() => <IntroCreationLazyPage telemetryService={telemetryService} />}
            />

            <CompatRoute
                path={`${match.url}/search`}
                render={() => (
                    <InsightCreationLazyPage
                        mode={InsightCreationPageType.Search}
                        telemetryService={telemetryService}
                    />
                )}
            />

            <CompatRoute
                path={`${match.url}/capture-group`}
                render={() => (
                    <InsightCreationLazyPage
                        mode={InsightCreationPageType.CaptureGroup}
                        telemetryService={telemetryService}
                    />
                )}
            />

            <CompatRoute
                path={`${match.url}/lang-stats`}
                exact={true}
                render={() => (
                    <InsightCreationLazyPage
                        mode={InsightCreationPageType.LangStats}
                        telemetryService={telemetryService}
                    />
                )}
            />

            {codeInsightsCompute && (
                <CompatRoute
                    path={`${match.url}/group-results`}
                    render={() => (
                        <InsightCreationLazyPage
                            mode={InsightCreationPageType.Compute}
                            telemetryService={telemetryService}
                        />
                    )}
                />
            )}
        </Switch>
    )
}
