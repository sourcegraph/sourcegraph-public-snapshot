import { Observable, of } from 'rxjs'
import { map, mapTo, switchMap } from 'rxjs/operators'
import { queryGraphQL } from '../../../../packages/webapp/src/backend/graphql'
import { gql, mutateGraphQL } from '../../../../packages/webapp/src/backend/graphql'
import { RegistryPublisher } from '../../../../packages/webapp/src/backend/graphqlschema'
import * as GQL from '../../../../packages/webapp/src/backend/graphqlschema'
import { createAggregateError } from '../../../../packages/webapp/src/util/errors'

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
