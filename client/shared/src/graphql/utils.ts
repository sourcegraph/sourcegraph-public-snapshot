import { ApolloQueryResult, ObservableQuery, Observable as ApolloObservable, gql as apolloGql } from '@apollo/client'
import type { DocumentNode } from 'graphql'
import { Observable, observable } from 'rxjs'

import type { GraphQLRequestDocument } from './graphql'

// Apollo's QueryObservable is not compatible with RxJS
// https://github.com/ReactiveX/rxjs/blob/9fb0ce9e09c865920cf37915cc675e3b3a75050b/src/internal/util/subscribeTo.ts#L32
export function fixObservable<T>(obz: ObservableQuery<T>): Observable<ApolloQueryResult<T>>
export function fixObservable<T>(obz: ApolloObservable<T>): Observable<T>
export function fixObservable<T>(
    obz: ApolloObservable<T> | ObservableQuery<T>
): Observable<ApolloQueryResult<T>> | Observable<T> {
    ;(obz as any)[observable] = () => obz
    return obz as any
}

/**
 * Returns a `DocumentNode` value to support integrations with GraphQL clients that require this.
 *
 * @param document The GraphQL operation payload
 * @returns The created `DocumentNode`
 */
export const getDocumentNode = (document: GraphQLRequestDocument): DocumentNode => {
    if (typeof document === 'string') {
        return apolloGql(document)
    }
    return document
}
