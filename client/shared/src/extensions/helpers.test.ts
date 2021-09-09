import { GraphQLError } from 'graphql'
import { of } from 'rxjs'
import { TestScheduler } from 'rxjs/testing'

import { ErrorGraphQLResult, SuccessGraphQLResult } from '../graphql/graphql'
import { PlatformContext } from '../platform/context'

import { ConfiguredExtension, ConfiguredExtensionManifestDefaultFields } from './extension'
import { ExtensionManifest } from './extensionManifest'
import { queryConfiguredRegistryExtensions } from './helpers'

const TEST_MANIFEST_RAW = '{"publisher":"a","url":"https://example.com","activationEvents":[]}'
const TEST_MANIFEST: Pick<ExtensionManifest, ConfiguredExtensionManifestDefaultFields | 'publisher'> = {
    publisher: 'a',
    url: 'https://example.com',
    activationEvents: [],
}

const scheduler = (): TestScheduler => new TestScheduler((actual, expected) => expect(actual).toStrictEqual(expected))

describe('queryConfiguredRegistryExtensions', () => {
    it('gets extensions from GraphQL servers supporting extensions(extensionIDs)', () => {
        const requestGraphQL: PlatformContext['requestGraphQL'] = () =>
            // eslint-disable-next-line @typescript-eslint/consistent-type-assertions
            of({
                data: {
                    extensionRegistry: {
                        extensions: {
                            nodes: [{ extensionID: 'a/b', manifest: { jsonFields: TEST_MANIFEST } }],
                        },
                    },
                },
            } as SuccessGraphQLResult<any>)
        scheduler().run(({ expectObservable }) => {
            expectObservable(queryConfiguredRegistryExtensions({ requestGraphQL }, ['a/b'])).toBe('(a|)', {
                a: [{ id: 'a/b', manifest: TEST_MANIFEST }] as ConfiguredExtension[],
            })
        })
    })

    it('gets extensions from GraphQL servers not supporting extensions(extensionIDs)/jsonFields and only supporting prioritizeExtensionIDs', () => {
        let calledWithExtensionIDsParameter = false
        const requestGraphQL: PlatformContext['requestGraphQL'] = ({ request }) => {
            if (request.includes('prioritizeExtensionIDs: ') && !request.includes('jsonFields(')) {
                // eslint-disable-next-line @typescript-eslint/consistent-type-assertions
                return of({
                    data: {
                        extensionRegistry: {
                            extensions: {
                                nodes: [{ extensionID: 'a/b', manifest: { raw: TEST_MANIFEST_RAW } }],
                            },
                        },
                    },
                } as SuccessGraphQLResult<any>)
            }
            calledWithExtensionIDsParameter = true
            // eslint-disable-next-line @typescript-eslint/consistent-type-assertions
            return of({
                data: undefined,
                errors: [
                    {
                        message: 'Unknown argument "extensionIDs" on field "extensions" of type "ExtensionRegistry".',
                    },
                    {
                        message: 'Cannot query field "jsonFields" on type "ExtensionManifest".',
                    },
                ] as GraphQLError[],
            } as ErrorGraphQLResult)
        }
        scheduler().run(({ expectObservable }) => {
            expectObservable(queryConfiguredRegistryExtensions({ requestGraphQL }, ['a/b'])).toBe('(a|)', {
                a: [{ id: 'a/b', manifest: TEST_MANIFEST }] as ConfiguredExtension[],
            })
        })
        expect(calledWithExtensionIDsParameter).toBeTruthy()
    })
})
