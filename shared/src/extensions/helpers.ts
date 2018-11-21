import { from, Observable, of, throwError } from 'rxjs'
import { catchError, filter, map, startWith, switchMap } from 'rxjs/operators'
import { gql, graphQLContent } from '../graphql/graphql'
import * as GQL from '../graphql/schema'
import { PlatformContext } from '../platform/context'
import { SettingsCascadeOrError } from '../settings/settings'
import { asError, createAggregateError, ErrorLike, isErrorLike } from '../util/errors'
import { ConfiguredExtension, toConfiguredExtension } from './extension'

const LOADING: 'loading' = 'loading'

export function viewerConfiguredExtensions({
    settingsCascade,
    queryGraphQL,
}: Pick<PlatformContext, 'settingsCascade' | 'queryGraphQL'>): Observable<ConfiguredExtension[]> {
    return viewerConfiguredExtensionsOrLoading({ settingsCascade, queryGraphQL }).pipe(
        filter((extensions): extensions is ConfiguredExtension[] | ErrorLike => extensions !== LOADING),
        switchMap(extensions => (isErrorLike(extensions) ? throwError(extensions) : [extensions]))
    )
}

function viewerConfiguredExtensionsOrLoading({
    settingsCascade,
    queryGraphQL,
}: Pick<PlatformContext, 'settingsCascade' | 'queryGraphQL'>): Observable<
    typeof LOADING | ConfiguredExtension[] | ErrorLike
> {
    return from(settingsCascade).pipe(
        switchMap(
            cascade =>
                isErrorLike(cascade.final)
                    ? [cascade.final]
                    : queryConfiguredExtensions({ queryGraphQL }, cascade).pipe(
                          catchError(error => [asError(error) as ErrorLike]),
                          startWith(LOADING)
                      )
        )
    )
}

export function queryConfiguredExtensions(
    { queryGraphQL }: Pick<PlatformContext, 'queryGraphQL'>,
    cascade: SettingsCascadeOrError
): Observable<ConfiguredExtension[]> {
    if (isErrorLike(cascade.final)) {
        return throwError(cascade.final)
    }
    if (!cascade.final || !cascade.final.extensions) {
        return of([])
    }
    const extensionIDs = Object.keys(cascade.final.extensions)
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
            },
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
            const configuredExtensions: ConfiguredExtension[] = []
            for (const extensionID of extensionIDs) {
                const registryExtension = registryExtensions.find(x => x.extensionID === extensionID)
                configuredExtensions.push(
                    registryExtension
                        ? toConfiguredExtension(registryExtension)
                        : { id: extensionID, manifest: null, rawManifest: null, registryExtension: undefined }
                )
            }
            return configuredExtensions
        })
    )
}
