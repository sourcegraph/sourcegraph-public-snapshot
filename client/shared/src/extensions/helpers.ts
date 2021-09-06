import { from, Observable, of } from 'rxjs'
import { map, switchMap } from 'rxjs/operators'

import { gql } from '../graphql/graphql'
import * as GQL from '../graphql/schema'
import { PlatformContext } from '../platform/context'
import { createAggregateError } from '../util/errors'

import {
    ConfiguredExtension,
    ConfiguredExtensionManifestDefaultFields,
    CONFIGURED_EXTENSION_DEFAULT_MANIFEST_FIELDS,
} from './extension'
import { parseExtensionManifestOrError, ExtensionManifest } from './extensionManifest'

/**
 * Query the GraphQL API for registry metadata about the extensions given in {@link extensionIDs}.
 *
 * @returns An observable that emits once with the results.
 */
export function queryConfiguredRegistryExtensions(
    // TODO(tj): can copy this over to extension host, just replace platformContext.requestGraphQL
    // with mainThreadAPI.requestGraphQL
    { requestGraphQL }: Pick<PlatformContext, 'requestGraphQL'>,
    extensionIDs: string[]
): Observable<ConfiguredExtension[]> {
    if (extensionIDs.length === 0) {
        return of([])
    }
    const variables: GQL.IExtensionsOnExtensionRegistryArguments = {
        first: extensionIDs.length,
        extensionIDs,
    }
    return from(
        requestGraphQL<GQL.IQuery>({
            request: gql`
                query Extensions($first: Int!, $extensionIDs: [String!]!, $extensionManifestFields: [String!]!) {
                    extensionRegistry {
                        extensions(first: $first, extensionIDs: $extensionIDs) {
                            nodes {
                                extensionID
                                manifest {
                                    jsonFields(fields: $extensionManifestFields)
                                }
                            }
                        }
                    }
                }
            `,
            variables: { ...variables, extensionManifestFields: CONFIGURED_EXTENSION_DEFAULT_MANIFEST_FIELDS },
            mightContainPrivateInfo: false,
        })
    ).pipe(
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
            return hasUnknownArgumentExtensionIDsError
                ? requestGraphQL<GQL.IQuery>({
                      request: gql`
                          query ExtensionsWithPrioritizeExtensionIDsParamAndNoJSONFields(
                              $first: Int!
                              $extensionIDs: [String!]!
                          ) {
                              extensionRegistry {
                                  extensions(first: $first, prioritizeExtensionIDs: $extensionIDs) {
                                      nodes {
                                          extensionID
                                          manifest {
                                              raw
                                          }
                                      }
                                  }
                              }
                          }
                      `,
                      variables,
                      mightContainPrivateInfo: false,
                  })
                : of({ data, errors })
        }),
        map(({ data, errors }) => {
            if (!data?.extensionRegistry?.extensions?.nodes) {
                throw createAggregateError(errors)
            }
            return data.extensionRegistry.extensions.nodes
                .filter(({ extensionID }) => extensionIDs.includes(extensionID))
                .map(({ extensionID, manifest }) => ({
                    id: extensionID,
                    manifest: manifest
                        ? (manifest.jsonFields as Pick<ExtensionManifest, ConfiguredExtensionManifestDefaultFields>) ||
                          parseExtensionManifestOrError(manifest.raw)
                        : null,
                }))
        })
    )
}
