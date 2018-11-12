import { from, Observable, of, throwError } from 'rxjs'
import { catchError, filter, map, startWith, switchMap } from 'rxjs/operators'
import { Context } from './context'
import { asError, createAggregateError, ErrorLike, isErrorLike } from './errors'
import { ConfiguredExtension } from './extensions/extension'
import { gql, graphQLContent, GraphQLDocument } from './graphql'
import * as GQL from './graphqlschema'
import { ExtensionManifest } from './schema/extension.schema'
import { Settings, SettingsCascadeOrError, SettingsSubject } from './settings'
import { parseJSONCOrError } from './util'

/**
 * A controller that exposes functionality for a settings cascade and querying extensions from the remote registry.
 */
export class Controller<S extends SettingsSubject, C extends Settings> {
    public static readonly LOADING: 'loading' = 'loading'

    constructor(public readonly context: Context<S, C>) {}

    private readonly viewerConfiguredExtensionsOrLoading: Observable<
        typeof Controller.LOADING | ConfiguredExtension[] | ErrorLike
    > = from(this.context.settingsCascade).pipe(
        switchMap(
            cascade =>
                isErrorLike(cascade.final)
                    ? [cascade.final]
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
            .pipe(map(registryExtension => this.withConfiguration([registryExtension])[0]))
    }

    public withRegistryMetadata(
        cascade: SettingsCascadeOrError<SettingsSubject, Settings>
    ): Observable<ConfiguredExtension[]> {
        if (isErrorLike(cascade.final)) {
            return throwError(cascade.final)
        }
        if (!cascade.final || !cascade.final.extensions) {
            return of([])
        }
        const extensionIDs = Object.keys(cascade.final.extensions)
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

    public withConfiguration(registryExtensions: GQL.IRegistryExtension[]): ConfiguredExtension[] {
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
    }
}
