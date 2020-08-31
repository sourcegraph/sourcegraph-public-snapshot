import * as H from 'history'
import { storiesOf } from '@storybook/react'
import { radios } from '@storybook/addon-knobs'
import React from 'react'
import webStyles from '../../SourcegraphWebApp.scss'
import { Tooltip } from '../tooltip/Tooltip'
import { ExternalServicePage } from './ExternalServicePage'
import { NOOP_TELEMETRY_SERVICE } from '../../../../shared/src/telemetry/telemetryService'
import { fetchExternalService as _fetchExternalService } from './backend'
import { of } from 'rxjs'
import { ExternalServiceKind } from '../../graphql-operations'

let isLightTheme = true
const { add } = storiesOf('web/External services/ExternalServicePage', module)
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

const fetchExternalService: typeof _fetchExternalService = () =>
    of({
        id: 'service123',
        kind: ExternalServiceKind.GITHUB,
        warning: null,
        config: '{"githubconfig": true}',
        displayName: 'GitHub.com',
        webhookURL: null,
    })

add('View external service config', () => {
    const history = H.createMemoryHistory()
    return (
        <ExternalServicePage
            history={history}
            afterUpdateRoute="/site-admin/after"
            telemetryService={NOOP_TELEMETRY_SERVICE}
            isLightTheme={isLightTheme}
            externalServiceID="service123"
            fetchExternalService={fetchExternalService}
            autoFocusForm={false}
        />
    )
})
