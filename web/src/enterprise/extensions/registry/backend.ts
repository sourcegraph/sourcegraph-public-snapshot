import { Observable, of } from 'rxjs'
import { map, mapTo, switchMap } from 'rxjs/operators'
import { RegistryPublisher } from '../../../../../shared/src/graphqlschema'
import * as GQL from '../../../../../shared/src/graphqlschema'
import { queryGraphQL } from '../../../backend/graphql'
import { gql, mutateGraphQL } from '../../../backend/graphql'
import { createAggregateError } from '../../../util/errors'

export function deleteRegistryExtensionWithConfirmation(extension: GQL.ID): Observable<boolean> {
    return of(window.confirm('Really delete this extension from the extension registry?')).pipe(
        switchMap(ok => {
            if (!ok) {
                return [false]
            }
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
                }),
                mapTo(true)
            )
        })
    )
}

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
