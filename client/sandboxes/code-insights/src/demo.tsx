import React, { ReactElement, Suspense } from 'react'
import { render } from 'react-dom'
import { BrowserRouter } from 'react-router-dom'
import { noop } from 'rxjs';

import { setLinkComponent } from '@sourcegraph/shared/src/components/Link'
import { SearchPatternType } from '@sourcegraph/shared/src/graphql-operations'
import { EMPTY_SETTINGS_CASCADE } from '@sourcegraph/shared/src/settings/settings'
import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { RouterLinkOrAnchor } from '@sourcegraph/web/src/components/RouterLinkOrAnchor'
import { InsightsApiContext } from '@sourcegraph/web/src/insights'
import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'

import { MockInsightsApi } from './mock-api'

import '@sourcegraph/web/src/SourcegraphWebApp.scss'
import { lazyComponent } from '@sourcegraph/web/src/util/lazyComponent';
import { Route, Switch } from 'react-router';
import { LayoutRouteProps } from '@sourcegraph/web/src/routes';

const mockAPI = new MockInsightsApi()

const CONTEXT = {
    versionContext: undefined,
    telemetryService: NOOP_TELEMETRY_SERVICE,
    copyQueryButton: false,
    setCaseSensitivity: noop,
    setPatternType: noop,
    patternType: SearchPatternType.literal,
    settingsCascade: EMPTY_SETTINGS_CASCADE,
    globbing: false,
    extensionsController: null,
}

/**
 * Mocking behavior and logic for rendering routes like Layout.tsx component does
 * for dev demo purposes.
 * */
const routes: readonly LayoutRouteProps<any>[] = [
    {
        path: '/insights',
        render: lazyComponent(() => import('@sourcegraph/web/src/insights/InsightsRouter'), 'InsightsRouter'),
    },
]

setLinkComponent(RouterLinkOrAnchor)

/** Main entry point to code-insights demo */
export function App(): ReactElement {
    return (
        <BrowserRouter>
            <InsightsApiContext.Provider value={mockAPI}>
                <Suspense
                    fallback={
                        <div className="flex flex-1">
                            <LoadingSpinner className="icon-inline m-2" />
                        </div>
                    }
                >
                    <Switch>
                        {/* eslint-disable react/jsx-no-bind */}
                        {routes.map(
                            ({ render, ...route }) =>
                                <Route
                                    {...route}
                                    key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                                    component={undefined}
                                    render={routeComponentProps => (
                                        <div className="layout__app-router-container">
                                            {/*// @ts-ignore*/}
                                            {render({ ...CONTEXT, ...routeComponentProps })}
                                        </div>
                                    )}
                                />
                        )}
                        {/* eslint-enable react/jsx-no-bind */}
                    </Switch>
                </Suspense>
            </InsightsApiContext.Provider>
        </BrowserRouter>
    )
}

render(<App />, document.querySelector('#root'))
