import { flatten } from 'lodash'
import React from 'react'
import renderer from 'react-test-renderer'
import { RegistryExtensionOverviewPage } from './RegistryExtensionOverviewPage'

describe('RegistryExtensionOverviewPage', () => {
    describe('categories', () => {
        test('filters out unrecognized categories', () => {
            const x = renderer.create(
                <RegistryExtensionOverviewPage
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
            ).root
            expect(
                flatten(
                    x
                        .findAll(
                            ({ props: { className } }) =>
                                className && className.includes('registry-extension-overview-page__categories')
                        )
                        .map(c => flatten(c.children.map(c => (typeof c === 'string' ? c : c.children))))
                )
            ).toEqual(['Other', 'Programming languages' /* no 'invalid' */])
        })
    })
})
