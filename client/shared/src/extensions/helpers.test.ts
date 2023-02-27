import { createGraphQLClientGetter } from '../testing/apollo/createGraphQLClientGetter'

import { ConfiguredExtension, ConfiguredExtensionManifestDefaultFields } from './extension'
import { ExtensionManifest } from './extensionManifest'
import { queryConfiguredRegistryExtensions } from './helpers'

const TEST_MANIFEST: Pick<ExtensionManifest, ConfiguredExtensionManifestDefaultFields | 'publisher'> = {
    publisher: 'a',
    url: 'https://example.com',
    activationEvents: [],
}

describe('queryConfiguredRegistryExtensions', () => {
    it('gets extensions from GraphQL servers supporting extensions(extensionIDs)', done => {
        const extensionsMock = {
            data: {
                extensionRegistry: {
                    extensions: {
                        nodes: [{ extensionID: 'a/b', manifest: { jsonFields: TEST_MANIFEST } }],
                    },
                },
            },
        }

        const getGraphQLClient = createGraphQLClientGetter({ watchQueryMocks: [extensionsMock] })

        queryConfiguredRegistryExtensions({ getGraphQLClient }, ['a/b']).subscribe(data => {
            expect(data).toEqual([{ id: 'a/b', manifest: TEST_MANIFEST }] as ConfiguredExtension[])
            done()
        })
    })
})
