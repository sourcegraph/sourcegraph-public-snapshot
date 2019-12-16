import { noop } from 'lodash'
import React from 'react'
import { MemoryRouter } from 'react-router'
import renderer from 'react-test-renderer'
import { RegistryExtensionOverviewPage } from './RegistryExtensionOverviewPage'

describe('RegistryExtensionOverviewPage', () => {
    test('renders', () =>
        expect(
            renderer
                .create(
                    <MemoryRouter>
                        <RegistryExtensionOverviewPage
                            eventLogger={{ logViewEvent: noop }}
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
                        />
                    </MemoryRouter>
                )
                .toJSON()
        ).toMatchSnapshot())

    describe('categories', () => {
        test('filters out unrecognized categories', () => {
            const x = renderer.create(
                <MemoryRouter>
                    <RegistryExtensionOverviewPage
                        eventLogger={{ logViewEvent: noop }}
                        extension={{
                            id: 'x',
                            rawManifest: '',
                            manifest: {
                                url: 'https://example.com',
                                activationEvents: ['*'],
                                categories: ['Programming languages', 'invalid', 'Other'],
                            },
                        }}
                    />
                </MemoryRouter>
            ).root
            expect(
                toText(
                    x.findAll(({ props: { className } }) =>
                        className?.includes('registry-extension-overview-page__categories')
                    )
                )
            ).toEqual(['Other', 'Programming languages' /* no 'invalid' */])
        })
    })
})

function toText(values: (string | renderer.ReactTestInstance)[]): string[] {
    const textNodes: string[] = []
    for (const v of values) {
        if (typeof v === 'string') {
            textNodes.push(v)
        } else {
            textNodes.push(...toText(v.children))
        }
    }
    return textNodes
}
