import { from, Observable, of, throwError } from 'rxjs'
import { catchError, filter, map, startWith, switchMap } from 'rxjs/operators'
import { Context } from './context'
import { asError, createAggregateError, ErrorLike, isErrorLike } from './errors'
import { ConfiguredExtension, toConfiguredExtension } from './extensions/extension'
import { gql, graphQLContent } from './graphql'
import * as GQL from './graphqlschema'
import { SettingsCascadeOrError } from './settings'

const LOADING: 'loading' = 'loading'

export function viewerConfiguredExtensions({
    settingsCascade,
    queryGraphQL,
}: Pick<Context, 'settingsCascade' | 'queryGraphQL'>): Observable<ConfiguredExtension[]> {
    return viewerConfiguredExtensionsOrLoading({ settingsCascade, queryGraphQL }).pipe(
        filter((extensions): extensions is ConfiguredExtension[] | ErrorLike => extensions !== LOADING),
        switchMap(extensions => (isErrorLike(extensions) ? throwError(extensions) : [extensions]))
    )
}

function viewerConfiguredExtensionsOrLoading({
    settingsCascade,
    queryGraphQL,
}: Pick<Context, 'settingsCascade' | 'queryGraphQL'>): Observable<typeof LOADING | ConfiguredExtension[] | ErrorLike> {
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
    { queryGraphQL }: Pick<Context, 'queryGraphQL'>,
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
