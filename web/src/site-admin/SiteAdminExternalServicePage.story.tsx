import { storiesOf } from '@storybook/react'
import { radios } from '@storybook/addon-knobs'
import React from 'react'
import { SiteAdminExternalServicePage } from './SiteAdminExternalServicePage'
import { createMemoryHistory } from 'history'
import webStyles from '../SourcegraphWebApp.scss'
import { MemoryRouter } from 'react-router'
import { NOOP_TELEMETRY_SERVICE } from '../../../shared/src/telemetry/telemetryService'
import { action } from '@storybook/addon-actions'
import { of } from 'rxjs'
import { IExternalService, ExternalServiceKind } from '../../../shared/src/graphql/schema'
import { map, tap } from 'rxjs/operators'

let isLightTheme = true

const { add } = storiesOf('web/site-admin/SiteAdminExternalServicePage', module).addDecorator(story => {
    const theme = radios('Theme', { Light: 'light', Dark: 'dark' }, 'light')
    document.body.classList.toggle('theme-light', theme === 'light')
    document.body.classList.toggle('theme-dark', theme === 'dark')
    isLightTheme = theme === 'light'
    return (
        <MemoryRouter>
            <style>{webStyles}</style>
            <div className="p-3 container">{story()}</div>
        </MemoryRouter>
    )
})

const hosts: Record<string, IExternalService> = {
    GitHub: {
        __typename: 'ExternalService',
        id: 'RXh0ZXJuYWxTZXJ2aWNlOjQ=',
        kind: ExternalServiceKind.GITHUB,
        displayName: 'GitHub Public',
        config:
            '{\n  "url": "https://github.com",\n  "token": "XXX",\n  // Sync no repositories. We only want the token to raise rate limits.\n  "repositoryQuery": [\n    "none"\n  ],\n  "exclude": [\n    {\n      "id": "MDEwOlJlcG9zaXRvcnkxOTI2MDUxODY=",\n      "name": "creachadair/jrpc2"\n    }\n  ]\n}',
        warning: null,
        webhookURL: null,
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
    },
    GitLab: {
        __typename: 'ExternalService',
        id: 'RXh0ZXJuYWxTZXJ2aWNlOjg=',
        kind: ExternalServiceKind.GITLAB,
        displayName: 'GitLab.com',
        config:
            '{\n  "url": "https://gitlab.com",\n  "token": "XXX",\n  // Sync no repositories. We only want the token to raise rate limits.\n  "projectQuery": [\n    "none",\n    "groups/sourcegraph/projects"\n  ]\n}',
        warning: null,
        webhookURL: null,
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
    },
}

for (const name of Object.keys(hosts).sort()) {
    add(name, () => {
        const history = createMemoryHistory()
        return (
            <SiteAdminExternalServicePage
                history={history}
                location={history.location}
                isLightTheme={isLightTheme}
                match={{
                    isExact: true,
                    path: '/site-admin/external/FOO',
                    url: 'http://test.test/site-admin/external/FOO',
                    params: { id: 'FOO' },
                }}
                fetchExternalService={() => of(hosts[name])}
                updateExternalService={() =>
                    of(undefined).pipe(
                        tap(() => action('Update external service')),
                        map(() => hosts[name])
                    )
                }
                telemetryService={NOOP_TELEMETRY_SERVICE}
            />
        )
    })
}
