import { createMemoryHistory } from 'history'
import React from 'react'
import { Router } from 'react-router'
import renderer from 'react-test-renderer'

import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { PageTitle } from '../../components/PageTitle'

import { RegistryExtensionOverviewPage } from './RegistryExtensionOverviewPage'

jest.mock('mdi-react/GithubIcon', () => 'GithubIcon')

describe('RegistryExtensionOverviewPage', () => {
    afterEach(() => {
        PageTitle.titleSet = false
    })
    test('renders', () => {
        const history = createMemoryHistory()
        expect(
            renderer
                .create(
                    <Router history={history}>
                        <RegistryExtensionOverviewPage
                            telemetryService={NOOP_TELEMETRY_SERVICE}
                            extension={{
                                id: 'x',
                                rawManifest: '{}',
                                manifest: {
                                    url: 'https://example.com',
                                    activationEvents: ['*'],
                                    categories: ['Programming languages', 'Other'],
                                    tags: ['T1', 'T2'],
                                    readme: '**A**',
                                    repository: {
                                        url: 'https://github.com/foo/bar',
                                        type: 'git',
                                    },
                                },
                            }}
                            history={history}
                            isLightTheme={true}
                        />
                    </Router>
                )
                .toJSON()
        ).toMatchSnapshot()
    })

    // Meant to prevent this regression: https://github.com/sourcegraph/sourcegraph/pull/15040
    test('renders with no tags', () => {
        const history = createMemoryHistory()
        expect(
            renderer
                .create(
                    <Router history={history}>
                        <RegistryExtensionOverviewPage
                            telemetryService={NOOP_TELEMETRY_SERVICE}
                            extension={{
                                id: 'x',
                                rawManifest: '{}',
                                manifest: {
                                    url: 'https://example.com',
                                    activationEvents: ['*'],
                                    categories: ['Programming languages', 'Other'],
                                    tags: [],
                                    readme: '**A**',
                                    repository: {
                                        url: 'https://github.com/foo/bar',
                                        type: 'git',
                                    },
                                },
                            }}
                            history={history}
                            isLightTheme={true}
                        />
                    </Router>
                )
                .toJSON()
        ).toMatchSnapshot()
    })

    describe('categories', () => {
        test('filters out unrecognized categories', () => {
            const history = createMemoryHistory()
            const output = renderer.create(
                <Router history={history}>
                    <RegistryExtensionOverviewPage
                        telemetryService={NOOP_TELEMETRY_SERVICE}
                        extension={{
                            id: 'x',
                            rawManifest: '',
                            manifest: {
                                url: 'https://example.com',
                                activationEvents: ['*'],
                                categories: ['Programming languages', 'invalid', 'Other'],
                            },
                        }}
                        history={createMemoryHistory()}
                        isLightTheme={true}
                    />
                </Router>
            ).root
            expect(
                toText(
                    output.findAll(({ props: { className } }) =>
                        className?.includes('test-registry-extension-categories')
                    )
                )
            ).toEqual(['Programming languages', 'Other' /* no 'invalid' */])
        })
    })
})

function toText(values: (string | renderer.ReactTestInstance)[]): string[] {
    const textNodes: string[] = []
    for (const value of values) {
        if (typeof value === 'string') {
            textNodes.push(value)
        } else {
            textNodes.push(...toText(value.children))
        }
    }
    return textNodes
}
