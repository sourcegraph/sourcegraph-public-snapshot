import { combineLatest, from, Observable, of, throwError } from 'rxjs'
import { catchError, filter, map, startWith, switchMap } from 'rxjs/operators'
import { Context } from './context'
import { asError, createAggregateError, ErrorLike, isErrorLike } from './errors'
import { ConfiguredExtension } from './extensions/extension'
import { gql, graphQLContent, GraphQLDocument } from './graphql'
import { ExtensionManifest } from './schema/extension.schema'
import * as GQL from './schema/graphqlschema'
import { ConfigurationCascadeOrError, ConfigurationSubject, Settings } from './settings'
import { parseJSONCOrError } from './util'

/**
 * A controller that exposes functionality for a configuration cascade and querying extensions from the remote
 * registry.
 */
export class Controller<S extends ConfigurationSubject, C extends Settings> {
    public static readonly LOADING: 'loading' = 'loading'

    constructor(public readonly context: Context<S, C>) {}

    private readonly viewerConfiguredExtensionsOrLoading: Observable<
        typeof Controller.LOADING | ConfiguredExtension[] | ErrorLike
    > = from(this.context.configurationCascade).pipe(
        switchMap(
            cascade =>
                isErrorLike(cascade.merged)
                    ? [cascade.merged]
                    : this.withRegistryMetadata(cascade).pipe(
                          catchError(error => [asError(error) as ErrorLike]),
                          startWith(Controller.LOADING)
                      )
        )
    )

    public readonly viewerConfiguredExtensions: Observable<
        ConfiguredExtension[]
    > = this.viewerConfiguredExtensionsOrLoading.pipe(
        filter((extensions): extensions is ConfiguredExtension[] | ErrorLike => extensions !== Controller.LOADING),
        switchMap(extensions => (isErrorLike(extensions) ? throwError(extensions) : [extensions]))
    )

    public forExtensionID(
        extensionID: string,
        registryExtensionFragment: GraphQLDocument | string
    ): Observable<ConfiguredExtension> {
        return from(
            this.context.queryGraphQL(
                gql`
                    query RegistryExtension($extensionID: String!) {
                        extensionRegistry {
                            extension(extensionID: $extensionID) {
                                ...RegistryExtensionFields
                            }
                        }
                    }
                    ${registryExtensionFragment}
                `[graphQLContent],
                { extensionID },
                false
            )
        )
            .pipe(
                map(({ data, errors }) => {
                    if (!data || !data.extensionRegistry || !data.extensionRegistry.extension) {
                        throw createAggregateError(errors)
                    }
                    return data.extensionRegistry.extension
                })
            )
            .pipe(
                switchMap(registryExtension => this.withConfiguration(of([registryExtension]))),
                map(configuredExtensions => configuredExtensions[0])
            )
    }

    public withRegistryMetadata(
        cascade: ConfigurationCascadeOrError<ConfigurationSubject, Settings>
    ): Observable<ConfiguredExtension[]> {
        if (isErrorLike(cascade.merged)) {
            return throwError(cascade.merged)
        }
        if (!cascade.merged || !cascade.merged.extensions) {
            return of([])
        }
        const extensionIDs = Object.keys(cascade.merged.extensions)
        return from(
            this.context.queryGraphQL(
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
                    configuredExtensions.push({
                        id: extensionID,
                        manifest:
                            registryExtension && registryExtension.manifest
                                ? parseJSONCOrError(registryExtension.manifest.raw)
                                : null,
                        rawManifest:
                            (registryExtension && registryExtension.manifest && registryExtension.manifest.raw) || null,
                        registryExtension,
                    })
                }
                return configuredExtensions
            })
        )
    }

    public withConfiguration(
        registryExtensions: Observable<GQL.IRegistryExtension[]>
    ): Observable<ConfiguredExtension[]> {
        return combineLatest(registryExtensions, this.context.configurationCascade).pipe(
            map(([registryExtensions, cascade]) => {
                const configuredExtensions: ConfiguredExtension[] = []
                for (const registryExtension of registryExtensions) {
                    configuredExtensions.push({
                        id: registryExtension.extensionID,
                        manifest: registryExtension.manifest
                            ? parseJSONCOrError<ExtensionManifest>(registryExtension.manifest.raw)
                            : null,
                        rawManifest:
                            (registryExtension && registryExtension.manifest && registryExtension.manifest.raw) || null,
                        registryExtension,
                    })
                }
                return configuredExtensions
            })
        )
    }
}
