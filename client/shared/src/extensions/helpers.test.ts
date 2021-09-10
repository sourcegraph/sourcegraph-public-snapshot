import { ApolloClient, NormalizedCacheObject, ObservableQuery } from '@apollo/client'
import { mock } from 'jest-mock-extended'
import { of, Subscriber } from 'rxjs'

import { ConfiguredExtension, ConfiguredExtensionManifestDefaultFields } from './extension'
import { ExtensionManifest } from './extensionManifest'
import { queryConfiguredRegistryExtensions } from './helpers'

const TEST_MANIFEST_RAW = '{"publisher":"a","url":"https://example.com","activationEvents":[]}'
const TEST_MANIFEST: Pick<ExtensionManifest, ConfiguredExtensionManifestDefaultFields | 'publisher'> = {
    publisher: 'a',
    url: 'https://example.com',
    activationEvents: [],
}

interface CreateGraphQLClientGetterOptions {
    /** Responses emitted by watch query sequentially for each call. */
    watchQueryMocks: object[]
}

/**
 * Creates Apollo Client getter mock.
 */
function createGraphQLClientGetter({ watchQueryMocks }: CreateGraphQLClientGetterOptions) {
    const observableQuery = mock<ObservableQuery<unknown, unknown>>()
    const graphQlClient = mock<ApolloClient<NormalizedCacheObject>>()

    graphQlClient.watchQuery.mockReturnValue(observableQuery)

    for (const mockResponse of watchQueryMocks) {
        observableQuery.subscribe.mockImplementationOnce((kek: unknown) =>
            of(mockResponse).subscribe(kek as Subscriber<unknown>)
        )
    }

    return () => Promise.resolve(graphQlClient)
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

    it('gets extensions from GraphQL servers not supporting extensions(extensionIDs)/jsonFields and only supporting prioritizeExtensionIDs', done => {
        const extensionsMock = {
            data: undefined,
            errors: [
                {
                    message: 'Unknown argument "extensionIDs" on field "extensions" of type "ExtensionRegistry".',
                },
                {
                    message: 'Cannot query field "jsonFields" on type "ExtensionManifest".',
                },
            ],
        }
        const extensionsCompatMock = {
            data: {
                extensionRegistry: {
                    extensions: {
                        nodes: [{ extensionID: 'a/b', manifest: { raw: TEST_MANIFEST_RAW } }],
                    },
                },
            },
        }

        const getGraphQLClient = createGraphQLClientGetter({ watchQueryMocks: [extensionsMock, extensionsCompatMock] })

        queryConfiguredRegistryExtensions({ getGraphQLClient }, ['a/b']).subscribe(data => {
            expect(data).toEqual([{ id: 'a/b', manifest: TEST_MANIFEST }] as ConfiguredExtension[])
            done()
        })
    })
})
