import { storiesOf } from '@storybook/react'
import { radios, boolean } from '@storybook/addon-knobs'
import React from 'react'
import * as GQL from '../../../shared/src/graphql/schema'
import { SiteAdminExternalServicePage } from './SiteAdminExternalServicePage'
import { createMemoryHistory } from 'history'
import webStyles from '../SourcegraphWebApp.scss'
import { MemoryRouter } from 'react-router'
import fetchMock from 'fetch-mock'

const { add } = storiesOf('web/site-admin/SiteAdminExternalServicePage', module).addDecorator(story => {
    const theme = radios('Theme', { Light: 'light', Dark: 'dark' }, 'light')
    document.body.classList.toggle('theme-light', theme === 'light')
    document.body.classList.toggle('theme-dark', theme === 'dark')
    return (
        <MemoryRouter>
            <style>{webStyles}</style>
            <div className="p-3 container">{story()}</div>
        </MemoryRouter>
    )
})

const hosts: { [key: string]: object } = {
    GitHub: {
        data: {
            node: {
                __typename: 'ExternalService',
                id: 'RXh0ZXJuYWxTZXJ2aWNlOjQ=',
                kind: 'GITHUB',
                displayName: 'GitHub Public',
                config:
                    '{\n  "url": "https://github.com",\n  // This token is from sourcegraph-dotcom-bot. It only has the public repo scope.\n  // IMPORTANT: If you change the token, you need to restart repo-updater\n  "token": "XXX",\n  // Sync no repositories. We only want the token to raise rate limits.\n  "repositoryQuery": [\n    "none"\n  ],\n  "exclude": [\n    {\n      "id": "MDEwOlJlcG9zaXRvcnkxOTI2MDUxODY=",\n      "name": "creachadair/jrpc2"\n    }\n  ]\n}',
                warning: null,
                webhookURL: null,
            },
        },
    },
    GitLab: {
        data: {
            node: {
                __typename: 'ExternalService',
                id: 'RXh0ZXJuYWxTZXJ2aWNlOjg=',
                kind: 'GITLAB',
                displayName: 'GitLab.com',
                config:
                    '{\n  "url": "https://gitlab.com",\n  // This token is from sourcegraph-dotcom-bot.\n  // IMPORTANT: If you change the token, you need to restart repo-updater\n  "token": "XXX",\n  // Sync no repositories. We only want the token to raise rate limits.\n  "projectQuery": [\n    "none",\n    "groups/sourcegraph/projects"\n  ]\n}',
                warning: null,
                webhookURL: null,
            },
        },
    },
}

for (const name of Object.keys(hosts).sort()) {
    add(name, () => {
        fetchMock
            .restore()
            .post('/.api/graphql?ExternalService', {
                body: hosts[name],
                status: 200,
            })
            .post('begin:/.api/graphql', {
                body: { data: { logUserEvent: null } },
                status: 200,
            })

        return (
            <SiteAdminExternalServicePage
                history={createMemoryHistory()}
                location={createMemoryHistory().location}
                isLightTheme={true}
                match={{
                    isExact: true,
                    path: '/site-admin/external/FOO',
                    url: 'http://test.test/site-admin/external/FOO',
                    params: { id: 'FOO' },
                }}
                mode="edit"
                input={{
                    kind: GQL.ExternalServiceKind.GITLAB,
                    displayName: 'GitLab',
                    config: '{}',
                }}
            />
        )
    })
}
