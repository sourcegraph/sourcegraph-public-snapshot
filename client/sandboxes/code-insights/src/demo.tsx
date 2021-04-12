import { createBrowserHistory } from 'history'
import React, { ReactElement, useState } from 'react'
import { render } from 'react-dom'
import { BrowserRouter } from 'react-router-dom'

import { setLinkComponent } from '@sourcegraph/shared/src/components/Link'
import { SearchPatternType } from '@sourcegraph/shared/src/graphql-operations'
import { EMPTY_SETTINGS_CASCADE } from '@sourcegraph/shared/src/settings/settings'
import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { RouterLinkOrAnchor } from '@sourcegraph/web/src/components/RouterLinkOrAnchor'
import { InsightsApiContext, InsightsPage } from '@sourcegraph/web/src/insights'

import { MockInsightsApi } from './mock-api'

import '@sourcegraph/web/src/SourcegraphWebApp.scss'

const history = createBrowserHistory()
const mockAPI = new MockInsightsApi()

setLinkComponent(RouterLinkOrAnchor)

export function App(): ReactElement {
    const [patternType, setPatterType] = useState(SearchPatternType.literal)
    const [caseSensitive, setCaseSensitive] = useState(false)

    return (
        <BrowserRouter>
            <InsightsApiContext.Provider value={mockAPI}>
                <InsightsPage
                    versionContext={undefined}
                    telemetryService={NOOP_TELEMETRY_SERVICE}
                    copyQueryButton={false}
                    caseSensitive={caseSensitive}
                    setCaseSensitivity={setCaseSensitive}
                    setPatternType={setPatterType}
                    patternType={patternType}
                    settingsCascade={EMPTY_SETTINGS_CASCADE}
                    globbing={false}
                    location={history.location}
                    history={history}
                    /* eslint-disable-next-line @typescript-eslint/ban-ts-comment */
                    // @ts-ignore
                    extensionsController={null}
                />
            </InsightsApiContext.Provider>
        </BrowserRouter>
    )
}

render(<App />, document.querySelector('#root'))
