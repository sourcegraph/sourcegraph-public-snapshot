import { Observable, of } from 'rxjs'
import { map, mapTo, switchMap } from 'rxjs/operators'

import { createAggregateError } from '@sourcegraph/common'
import { gql } from '@sourcegraph/http-client'
import * as GQL from '@sourcegraph/shared/src/schema'

import { mutateGraphQL, queryGraphQL, requestGraphQL } from '../../../backend/graphql'
import { DeleteRegistryExtensionResult, DeleteRegistryExtensionVariables, Scalars } from '../../../graphql-operations'

export function deleteRegistryExtensionWithConfirmation(extension: Scalars['ID']): Observable<boolean> {
    return of(window.confirm('Really delete this extension from the extension registry?')).pipe(
        switchMap(wasConfirmed => {
            if (!wasConfirmed) {
                return [false]
            }
            return requestGraphQL<DeleteRegistryExtensionResult, DeleteRegistryExtensionVariables>(
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
                    if (!data?.extensionRegistry?.deleteExtension || (errors && errors.length > 0)) {
                        throw createAggregateError(errors)
                    }
                }),
                mapTo(true)
            )
        })
    )
}

export function queryViewerRegistryPublishers(): Observable<GQL.RegistryPublisher[]> {
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
            if (!data?.extensionRegistry?.viewerPublishers || (errors && errors.length > 0)) {
                throw createAggregateError(errors)
            }
            return data.extensionRegistry.viewerPublishers.map(publisher => ({
                ...publisher,
                extensionIDPrefix: data.extensionRegistry.localExtensionIDPrefix || undefined,
            }))
        })
    )
}

export function createExtension(
    publisher: Scalars['ID'],
    name: string
): Observable<GQL.IExtensionRegistryCreateExtensionResult> {
    return mutateGraphQL(
        gql`
            mutation CreateRegistryExtension($publisher: ID!, $name: String!) {
                extensionRegistry {
                    createExtension(publisher: $publisher, name: $name) {
                        extension {
                            id
                            extensionID
                            url
                        }
                    }
                }
            }
        `,
        { publisher, name }
    ).pipe(
        map(({ data, errors }) => {
            if (
                !data ||
                !data.extensionRegistry ||
                !data.extensionRegistry.createExtension ||
                (errors && errors.length > 0)
            ) {
                throw createAggregateError(errors)
            }
            return data.extensionRegistry.createExtension
        })
    )
}
