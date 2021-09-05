import { from, Observable, of } from 'rxjs'
import { map } from 'rxjs/operators'

import { gql } from '../graphql/graphql'
import * as GQL from '../graphql/schema'
import { PlatformContext } from '../platform/context'
import { createAggregateError } from '../util/errors'

import { ConfiguredExtension } from './extension'
import { parseExtensionManifestOrError } from './extensionManifest'

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
        prioritizeExtensionIDs: extensionIDs,
    }
    return from(
        requestGraphQL<GQL.IQuery>({
            request: gql`
                query Extensions($first: Int!, $prioritizeExtensionIDs: [String!]!) {
                    extensionRegistry {
                        extensions(first: $first, prioritizeExtensionIDs: $prioritizeExtensionIDs) {
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
    ).pipe(
        map(({ data, errors }) => {
            if (!data?.extensionRegistry?.extensions?.nodes) {
                throw createAggregateError(errors)
            }
            return data.extensionRegistry.extensions.nodes
                .filter(({ extensionID }) => extensionIDs.includes(extensionID))
                .map<ConfiguredExtension>(({ extensionID, manifest }) => ({
                    id: extensionID,
                    manifest: manifest ? parseExtensionManifestOrError(manifest.raw) : null,
                    rawManifest: manifest?.raw ?? null,
                }))
        })
    )
}
