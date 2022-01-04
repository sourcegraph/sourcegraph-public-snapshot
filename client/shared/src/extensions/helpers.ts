import { Observable, of } from 'rxjs'
import { map, switchMap } from 'rxjs/operators'

import { createAggregateError } from '@sourcegraph/common'

import {
    ExtensionsResult,
    ExtensionsVariables,
    ExtensionsWithPrioritizeExtensionIDsParamAndNoJSONFieldsResult,
    ExtensionsWithPrioritizeExtensionIDsParamAndNoJSONFieldsVariables,
} from '../graphql-operations'
import { fromObservableQueryPromise, getDocumentNode, gql } from '../graphql/graphql'
import { PlatformContext } from '../platform/context'

import {
    ConfiguredExtension,
    ConfiguredExtensionManifestDefaultFields,
    CONFIGURED_EXTENSION_DEFAULT_MANIFEST_FIELDS,
} from './extension'
import { parseExtensionManifestOrError, ExtensionManifest } from './extensionManifest'

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

const ExtensionsWithPrioritizeExtensionIDsParameterAndNoJSONFieldsQuery = gql`
    query ExtensionsWithPrioritizeExtensionIDsParamAndNoJSONFields($first: Int!, $extensionIDs: [String!]!) {
        extensionRegistry {
            extensions(first: $first, prioritizeExtensionIDs: $extensionIDs) {
                nodes {
                    id
                    extensionID
                    manifest {
                        raw
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

    const variables: ExtensionsWithPrioritizeExtensionIDsParamAndNoJSONFieldsVariables = {
        first: extensionIDs.length,
        extensionIDs,
    }

    const queryObservablePromise = getGraphQLClient().then(client =>
        client.watchQuery<ExtensionsResult, ExtensionsVariables>({
            query: getDocumentNode(ExtensionsQuery),
            variables: {
                ...variables,
                // Spread operator is required to avoid Typescript type error
                // because of `readonly` type of `CONFIGURED_EXTENSION_DEFAULT_MANIFEST_FIELDS`.
                extensionManifestFields: [...CONFIGURED_EXTENSION_DEFAULT_MANIFEST_FIELDS],
            },
        })
    )

    return fromObservableQueryPromise(queryObservablePromise).pipe(
        switchMap(({ data, errors }) => {
            // BACKCOMPAT: The `extensionIDs` param to Query.extensionRegistry.extensions and the
            // ExtensionManifest#jsonFields field were added in 2021-09 and are not supported by
            // older Sourcegraph instances, so we need to catch the error and retry using the older
            // (and less-optimized) GraphQL query instead.
            const hasUnknownArgumentExtensionIDsError = errors?.some(
                error =>
                    error.message ===
                        'Unknown argument "extensionIDs" on field "extensions" of type "ExtensionRegistry".' ||
                    error.message === 'Cannot query field "jsonFields" on type "ExtensionManifest".'
            )

            if (!hasUnknownArgumentExtensionIDsError) {
                return of({ data, errors })
            }

            const queryObservablePromise = getGraphQLClient().then(client =>
                client.watchQuery<
                    ExtensionsWithPrioritizeExtensionIDsParamAndNoJSONFieldsResult,
                    ExtensionsWithPrioritizeExtensionIDsParamAndNoJSONFieldsVariables
                >({
                    query: getDocumentNode(ExtensionsWithPrioritizeExtensionIDsParameterAndNoJSONFieldsQuery),
                    variables,
                })
            )

            return fromObservableQueryPromise(queryObservablePromise)
        }),
        map(({ data, errors }) => {
            if (!data?.extensionRegistry?.extensions?.nodes) {
                throw createAggregateError(errors)
            }

            const { nodes } = data.extensionRegistry.extensions

            return (nodes as typeof nodes[number][])
                .filter(({ extensionID }) => extensionIDs.includes(extensionID))
                .map(({ extensionID, manifest }) => {
                    const getManifest = (value: typeof manifest): ConfiguredExtension['manifest'] => {
                        if (!value) {
                            return value
                        }

                        if ('jsonFields' in value) {
                            return value.jsonFields as Pick<ExtensionManifest, ConfiguredExtensionManifestDefaultFields>
                        }

                        return parseExtensionManifestOrError(value.raw)
                    }

                    return {
                        id: extensionID,
                        manifest: getManifest(manifest),
                    }
                })
        })
    )
}
