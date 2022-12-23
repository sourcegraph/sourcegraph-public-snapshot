import { Observable, of } from 'rxjs'
import { map } from 'rxjs/operators'

import { createAggregateError } from '@sourcegraph/common'
import { fromObservableQueryPromise, getDocumentNode, gql } from '@sourcegraph/http-client'

import { ExtensionsResult, ExtensionsVariables } from '../graphql-operations'
import { PlatformContext } from '../platform/context'

import {
    ConfiguredExtension,
    ConfiguredExtensionManifestDefaultFields,
    CONFIGURED_EXTENSION_DEFAULT_MANIFEST_FIELDS,
} from './extension'
import { ExtensionManifest } from './extensionManifest'

const ExtensionsQuery = gql`
    query Extensions($first: Int!, $extensionIDs: [String!]!, $extensionManifestFields: [String!]!) {
        extensionRegistry {
            extensions(first: $first, extensionIDs: $extensionIDs) {
                nodes {
                    id
                    extensionID
                    manifest {
                        jsonFields(fields: $extensionManifestFields)
                    }
                }
            }
        }
    }
`

/**
 * Query the GraphQL API for registry metadata about the extensions given in {@link extensionIDs}.
 *
 * @returns An observable that emits once with the results.
 */
export function queryConfiguredRegistryExtensions(
    // TODO(tj): can copy this over to extension host, just replace platformContext.requestGraphQL
    // with mainThreadAPI.requestGraphQL
    { getGraphQLClient }: Pick<PlatformContext, 'getGraphQLClient'>,
    extensionIDs: string[]
): Observable<ConfiguredExtension[]> {
    if (extensionIDs.length === 0) {
        return of([])
    }

    const queryObservablePromise = getGraphQLClient().then(client =>
        client.watchQuery<ExtensionsResult, ExtensionsVariables>({
            query: getDocumentNode(ExtensionsQuery),
            variables: {
                first: extensionIDs.length,
                extensionIDs,
                // Spread operator is required to avoid Typescript type error
                // because of `readonly` type of `CONFIGURED_EXTENSION_DEFAULT_MANIFEST_FIELDS`.
                extensionManifestFields: [...CONFIGURED_EXTENSION_DEFAULT_MANIFEST_FIELDS],
            },
        })
    )

    return fromObservableQueryPromise(queryObservablePromise).pipe(
        map(({ data, errors }) => {
            if (!data?.extensionRegistry?.extensions?.nodes) {
                throw createAggregateError(errors)
            }

            const { nodes } = data.extensionRegistry.extensions

            return nodes
                .filter(({ extensionID }) => extensionIDs.includes(extensionID))
                .map(({ extensionID, manifest }) => {
                    const getManifest = (value: typeof manifest): ConfiguredExtension['manifest'] => {
                        if (!value) {
                            return value
                        }

                        return value.jsonFields as Pick<ExtensionManifest, ConfiguredExtensionManifestDefaultFields>
                    }

                    return {
                        id: extensionID,
                        manifest: getManifest(manifest),
                    }
                })
        })
    )
}
