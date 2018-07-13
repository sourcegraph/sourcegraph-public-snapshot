import { Observable, of } from 'rxjs'
import { map, switchMap, take } from 'rxjs/operators'
import { gql, mutateGraphQL, queryGraphQL } from '../backend/graphql'
import * as GQL from '../backend/graphqlschema'
import { mutateConfigurationGraphQL } from '../configuration/backend'
import { configurationCascade } from '../settings/configuration'
import { createAggregateError } from '../util/errors'
import { RegistryPublisher } from './extension'

export function queryViewerRegistryPublishers(): Observable<RegistryPublisher[]> {
    return queryGraphQL(gql`
        query ViewerRegistryPublishers {
            extensionRegistry {
                viewerPublishers {
                    __typename
                    ... on User {
                        id
                        username
                    }
                    ... on Org {
                        id
                        name
                    }
                }
                localExtensionIDPrefix
            }
        }
    `).pipe(
        map(({ data, errors }) => {
            if (
                !data ||
                !data.extensionRegistry ||
                !data.extensionRegistry.viewerPublishers ||
                (errors && errors.length > 0)
            ) {
                throw createAggregateError(errors)
            }
            return data.extensionRegistry.viewerPublishers.map(p => ({
                ...p,
                extensionIDPrefix: data.extensionRegistry.localExtensionIDPrefix || undefined,
            }))
        })
    )
}

function updateExtensionSettings(
    subject: GQL.ConfigurationSubject | GQL.IConfigurationSubject | { id: GQL.ID },
    args: GQL.IUpdateExtensionOnConfigurationMutationArguments
): Observable<GQL.IUpdateExtensionConfigurationResult> {
    return mutateConfigurationGraphQL(
        subject,
        gql`
            mutation UpdateExtensionSettings(
                $subject: ID!
                $lastID: Int
                $extension: ID
                $extensionID: String
                $enabled: Boolean
                $remove: Boolean
                $edit: ConfigurationEdit
            ) {
                configurationMutation(input: { subject: $subject, lastID: $lastID }) {
                    updateExtension(
                        extension: $extension
                        extensionID: $extensionID
                        enabled: $enabled
                        remove: $remove
                        edit: $edit
                    ) {
                        mergedSettings
                    }
                }
            }
        `,
        args
    ).pipe(
        map(({ data, errors }) => {
            if (
                !data ||
                !data.configurationMutation ||
                !data.configurationMutation.updateExtension ||
                (errors && errors.length > 0)
            ) {
                throw createAggregateError(errors)
            }
            return data.configurationMutation.updateExtension
        })
    )
}

export function updateUserExtensionSettings(
    args: GQL.IUpdateExtensionOnConfigurationMutationArguments
): Observable<GQL.IUpdateExtensionConfigurationResult> {
    return configurationCascade.pipe(
        take(1),
        switchMap(configurationCascade =>
            updateExtensionSettings(
                // Only support configuring extension settings in user settings with this action.
                configurationCascade.subjects[configurationCascade.subjects.length - 1],
                args
            )
        )
    )
}

export function toGQLKeyPath(keyPath: (string | number)[]): GQL.IKeyPathSegment[] {
    return keyPath.map(v => (typeof v === 'string' ? { property: v } : { index: v }))
}

export function deleteRegistryExtensionWithConfirmation(extension: GQL.ID): Observable<void> {
    return of(window.confirm('Really delete this extension from the extension registry?')).pipe(
        switchMap(ok => {
            if (ok) {
                return mutateGraphQL(
                    gql`
                        mutation DeleteRegistryExtension($extension: ID!) {
                            extensionRegistry {
                                deleteExtension(extension: $extension) {
                                    alwaysNil
                                }
                            }
                        }
                    `,
                    { extension }
                ).pipe(
                    map(({ data, errors }) => {
                        if (
                            !data ||
                            !data.extensionRegistry ||
                            !data.extensionRegistry.deleteExtension ||
                            (errors && errors.length > 0)
                        ) {
                            throw createAggregateError(errors)
                        }
                    })
                )
            }
            return [void 0]
        })
    )
}
