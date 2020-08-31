import * as H from 'history'
import { storiesOf } from '@storybook/react'
import { radios } from '@storybook/addon-knobs'
import React from 'react'
import webStyles from '../../SourcegraphWebApp.scss'
import { Tooltip } from '../tooltip/Tooltip'
import { AddExternalServicesPage } from './AddExternalServicesPage'
import { NOOP_TELEMETRY_SERVICE } from '../../../../shared/src/telemetry/telemetryService'
import { fetchExternalService as _fetchExternalService } from './backend'
import { codeHostExternalServices, nonCodeHostExternalServices } from './externalServices'

let isLightTheme = true
const { add } = storiesOf('web/External services/AddExternalServicesPage', module)
    .addDecorator(story => {
        const theme = radios('Theme', { Light: 'light', Dark: 'dark' }, 'light')
        document.body.classList.toggle('theme-light', theme === 'light')
        document.body.classList.toggle('theme-dark', theme === 'dark')
        isLightTheme = theme === 'light'
        return (
            <>
                <Tooltip />
                <style>{webStyles}</style>
                <div className="p-3 container">{story()}</div>
            </>
        )
    })
    .addParameters({
        chromatic: {
            // Delay screenshot taking, so Monaco has some time to get syntax highlighting prepared.
            delay: 2000,
        },
    })

add('Overview', () => {
    const history = H.createMemoryHistory()
    return (
        <AddExternalServicesPage
            history={history}
            routingPrefix="/site-admin"
            telemetryService={NOOP_TELEMETRY_SERVICE}
            isLightTheme={isLightTheme}
            afterCreateRoute="/site-admin/after"
            codeHostExternalServices={codeHostExternalServices}
            nonCodeHostExternalServices={nonCodeHostExternalServices}
            autoFocusForm={false}
        />
    )
})

add('Add connection by kind', () => {
    const history = H.createMemoryHistory({ initialEntries: ['/page?id=github'] })
    return (
        <AddExternalServicesPage
            history={history}
            routingPrefix="/site-admin"
            telemetryService={NOOP_TELEMETRY_SERVICE}
            isLightTheme={isLightTheme}
            afterCreateRoute="/site-admin/after"
            codeHostExternalServices={codeHostExternalServices}
            nonCodeHostExternalServices={nonCodeHostExternalServices}
            autoFocusForm={false}
        />
    )
})
