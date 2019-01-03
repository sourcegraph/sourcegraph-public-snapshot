import { isEqual } from 'lodash'
import { from, Observable, of, throwError } from 'rxjs'
import {
    catchError,
    distinctUntilChanged,
    filter,
    map,
    publishReplay,
    refCount,
    startWith,
    switchMap,
} from 'rxjs/operators'
import { gql, graphQLContent } from '../graphql/graphql'
import * as GQL from '../graphql/schema'
import { PlatformContext } from '../platform/context'
import { asError, createAggregateError, ErrorLike, isErrorLike } from '../util/errors'
import { ConfiguredRegistryExtension, extensionIDsFromSettings, toConfiguredRegistryExtension } from './extension'

const LOADING: 'loading' = 'loading'

/**
 * @returns An observable that emits the list of extensions configured in the viewer's final settings upon
 * subscription and each time it changes.
 */
export function viewerConfiguredExtensions({
    settings,
    queryGraphQL,
}: Pick<PlatformContext, 'settings' | 'queryGraphQL'>): Observable<ConfiguredRegistryExtension[]> {
    return from(settings).pipe(
        map(settings => extensionIDsFromSettings(settings)),
        distinctUntilChanged((a, b) => isEqual(a, b)),
        switchMap(extensionIDs =>
            queryConfiguredRegistryExtensions({ queryGraphQL }, extensionIDs).pipe(startWith(LOADING))
        ),
        catchError(error => [asError(error) as ErrorLike]),
        filter((extensions): extensions is ConfiguredRegistryExtension[] | ErrorLike => extensions !== LOADING),
        switchMap(extensions => (isErrorLike(extensions) ? throwError(extensions) : [extensions])),
        publishReplay(),
        refCount()
    )
}

/**
 * Query the GraphQL API for registry metadata about the extensions given in {@link extensionIDs}.
 *
 * @returns An observable that emits once with the results.
 */
export function queryConfiguredRegistryExtensions(
    { queryGraphQL }: Pick<PlatformContext, 'queryGraphQL'>,
    extensionIDs: string[]
): Observable<ConfiguredRegistryExtension[]> {
    if (extensionIDs.length === 0) {
        return of([])
    }
    return from(
        queryGraphQL<GQL.IQuery>(
            gql`
                query Extensions($first: Int!, $prioritizeExtensionIDs: [String!]!) {
                    extensionRegistry {
                        extensions(first: $first, prioritizeExtensionIDs: $prioritizeExtensionIDs) {
                            nodes {
                                id
                                extensionID
                                url
                                manifest {
                                    raw
                                }
                                viewerCanAdminister
                            }
                        }
                    }
                }
            `[graphQLContent],
            {
                first: extensionIDs.length,
                prioritizeExtensionIDs: extensionIDs,
            } as GQL.IExtensionsOnExtensionRegistryArguments,
            false
        )
    ).pipe(
        map(({ data, errors }) => {
            if (
                !data ||
                !data.extensionRegistry ||
                !data.extensionRegistry.extensions ||
                !data.extensionRegistry.extensions.nodes
            ) {
                throw createAggregateError(errors)
            }
            return data.extensionRegistry.extensions.nodes.map(
                ({ id, extensionID, url, manifest, viewerCanAdminister }) => ({
                    id,
                    extensionID,
                    url,
                    manifest: manifest ? { raw: manifest.raw } : null,
                    viewerCanAdminister,
                })
            )
        }),
        map(registryExtensions => {
            const configuredExtensions: ConfiguredRegistryExtension[] = []
            for (const extensionID of extensionIDs) {
                const registryExtension = registryExtensions.find(x => x.extensionID === extensionID)
                configuredExtensions.push(
                    registryExtension
                        ? toConfiguredRegistryExtension(registryExtension)
                        : { id: extensionID, manifest: null, rawManifest: null, registryExtension: undefined }
                )
            }
            return configuredExtensions
        })
    )
}
