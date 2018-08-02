import { combineLatest, Observable, of, throwError } from 'rxjs'
import { catchError, filter, map, startWith, switchMap } from 'rxjs/operators'
import { Context } from './context'
import { Settings } from './copypasta'
import { asError, createAggregateError, ErrorLike, isErrorLike } from './errors'
import { ConfiguredExtension, isExtensionAdded, isExtensionEnabled } from './extensions/extension'
import { gql, graphQLContent, GraphQLDocument } from './graphql'
import { SourcegraphExtension } from './schema/extension.schema'
import * as GQL from './schema/graphqlschema'
import { ConfigurationCascade, ConfigurationSubject } from './settings'
import { parseJSONCOrError } from './util'

/**
 * A controller that exposes functionality for a configuration cascade and querying extensions from the remote
 * registry.
 */
export class Controller<S extends ConfigurationSubject, C = Settings> {
    public static readonly LOADING: 'loading' = 'loading'

    constructor(public readonly context: Context<S, C>) {}

    private readonly viewerConfiguredExtensionsOrLoading: Observable<
        typeof Controller.LOADING | ConfiguredExtension[] | ErrorLike
    > = this.context.configurationCascade.pipe(
        switchMap(
            cascade =>
                isErrorLike(cascade.merged)
                    ? [cascade.merged]
                    : this.withRegistryMetadata(cascade).pipe(
                          catchError(error => [asError(error)]),
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
        return this.context
            .queryGraphQL(
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
                { extensionID }
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
        cascade: ConfigurationCascade<ConfigurationSubject, Settings>
    ): Observable<ConfiguredExtension[]> {
        if (isErrorLike(cascade.merged)) {
            return throwError(cascade.merged)
        }
        if (!cascade.merged || !cascade.merged.extensions) {
            return of([])
        }
        const extensionIDs = Object.keys(cascade.merged.extensions)
        return this.context
            .queryGraphQL(
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
                }
            )
            .pipe(
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
                        const settingsProperties = toSettingsProperties(cascade, extensionID)
                        configuredExtensions.push({
                            extensionID,
                            ...settingsProperties,
                            manifest:
                                registryExtension && registryExtension.manifest
                                    ? parseJSONCOrError(registryExtension.manifest.raw)
                                    : null,
                            rawManifest:
                                (registryExtension && registryExtension.manifest && registryExtension.manifest.raw) ||
                                null,
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
                        extensionID: registryExtension.extensionID,
                        ...toSettingsProperties(cascade, registryExtension.extensionID),
                        manifest: registryExtension.manifest
                            ? parseJSONCOrError<SourcegraphExtension>(registryExtension.manifest.raw)
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

function toSettingsProperties(
    cascade: ConfigurationCascade<ConfigurationSubject, Settings>,
    extensionID: string
): Pick<ConfiguredExtension, 'settings' | 'settingsCascade' | 'isEnabled' | 'isAdded'> {
    const mergedSettings = cascade.merged
    return {
        settings: mergedSettings,
        settingsCascade: cascade.subjects.map(({ subject, settings }) => ({
            subject,
            settings,
        })),
        isEnabled: !isErrorLike(mergedSettings) && isExtensionEnabled(mergedSettings, extensionID),
        isAdded: !isErrorLike(mergedSettings) && isExtensionAdded(mergedSettings, extensionID),
    }
}
