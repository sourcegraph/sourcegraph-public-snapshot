import * as H from 'history'
import { storiesOf } from '@storybook/react'
import { radios } from '@storybook/addon-knobs'
import React from 'react'
import webStyles from '../../SourcegraphWebApp.scss'
import { Tooltip } from '../tooltip/Tooltip'
import { ExternalServicesPage } from './ExternalServicesPage'
import { NOOP_TELEMETRY_SERVICE } from '../../../../shared/src/telemetry/telemetryService'
import { queryExternalServices as _queryExternalServices } from './backend'
import { of } from 'rxjs'
import { ExternalServiceKind } from '../../graphql-operations'

const { add } = storiesOf('web/External services/ExternalServicesPage', module).addDecorator(story => {
    const theme = radios('Theme', { Light: 'light', Dark: 'dark' }, 'light')
    document.body.classList.toggle('theme-light', theme === 'light')
    document.body.classList.toggle('theme-dark', theme === 'dark')
    return (
        <>
            <Tooltip />
            <style>{webStyles}</style>
            <div className="p-3 container">{story()}</div>
        </>
    )
})

const queryExternalServices: typeof _queryExternalServices = () =>
    of({
        totalCount: 1,
        pageInfo: {
            endCursor: null,
            hasNextPage: false,
        },
        nodes: [
            {
                id: 'service1',
                kind: ExternalServiceKind.GITHUB,
                displayName: 'GitHub.com',
                config: '{"githubconfig":true}',
            },
        ],
    })

add('List of external services', () => {
    const history = H.createMemoryHistory()
    return (
        <ExternalServicesPage
            history={history}
            location={history.location}
            routingPrefix="/site-admin"
            afterDeleteRoute="/site-admin/after"
            telemetryService={NOOP_TELEMETRY_SERVICE}
            queryExternalServices={queryExternalServices}
        />
    )
})
