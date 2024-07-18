import {
    Client,
    cacheExchange,
    fetchExchange,
    makeOperation,
    mapExchange,
    type AnyVariables,
    type DocumentInput,
    type Exchange,
    type OperationResult,
} from '@urql/core'
import type { OperationDefinitionNode } from 'graphql'
import { once } from 'lodash'
import { type Readable, get, writable, type Writable } from 'svelte/store'

import type { GraphQLResult } from '@sourcegraph/http-client'

import { uniqueID } from '$lib/dom'
import { GRAPHQL_URI } from '$lib/http-client'

import { getHeaders } from './shared'

export type GraphQLClient = Client

export { gql, createRequest, type DocumentInput, type OperationResult } from '@urql/core'

/**
 * This exchange appends the operation name to the URL for each operation.
 */
const appendOperationName: Exchange = mapExchange({
    onOperation: op => {
        const operationName = op.query.definitions.find(
            (def): def is OperationDefinitionNode => def.kind === 'OperationDefinition'
        )?.name?.value
        if (operationName) {
            return makeOperation(op.kind, op, {
                ...op.context,
                url: `${op.context.url}?${operationName}`,
            })
        }
        return op
    },
})

export const getGraphQLClient = once((): Client => {
    return new Client({
        url: GRAPHQL_URI,
        fetchOptions: () => ({
            headers: getHeaders(),
        }),
        exchanges: [cacheExchange, appendOperationName, fetchExchange],
    })
})

// TODO: Refactor to eliminate the need for this function
/**
 * @deprecated Initiated GraphQL requests in data loader functions instead
 */
export function query<TData = any, TVariables extends AnyVariables = AnyVariables>(
    query: DocumentInput<TData, TVariables>,
    variables: TVariables
): Promise<OperationResult<TData, TVariables>> {
    return getGraphQLClient().query<TData, TVariables>(query, variables).toPromise()
}

interface InfinityQueryArgs<TData, TPayload = any, TVariables extends AnyVariables = AnyVariables, TSnapshot = any> {
    /**
     * The {@link Client} instance to use for the query.
     */
    client: Client

    /**
     * The GraphQL query to execute.
     */
    query: DocumentInput<TPayload, TVariables>

    /**
     * The initial variables to use for the query.
     */
    variables: TVariables | Promise<TVariables>

    /**
     * Process the result of the query. This function maps the response to the data used
     * and computes the next set of query variables, if any.
     *
     * @param result - The result of the query.
     * @param previousResult - The previous result of the query.
     * @returns The new/combined result state.
     */
    map: (result: OperationResult<TPayload, TVariables>) => InfinityStoreResult<TData, TVariables>

    /**
     * Optional callback to merge the data from the previous result with the new data.
     * If not provided the new data will replace the old data.
     */
    merge?: (previousData: TData | undefined, newData: TData | undefined) => TData

    /**
     * Returns a strategy for restoring the data when navigating back to a page.
     */
    createRestoreStrategy?: (api: InfinityAPI<TData, TVariables>) => RestoreStrategy<TSnapshot, TData>
}

/**
 * Internal API for the infinity query store. Used by restore strategies to control the store.
 */
interface InfinityAPI<TData, TVariables extends AnyVariables = AnyVariables> {
    /**
     * The internal store representing the current state of the query.
     */
    store: Writable<InfinityStoreResultState<TData, TVariables>>
    /**
     * Helper function for fetching and processing the next set of data.
     */
    fetch(
        variables: Partial<TVariables>,
        previous: InfinityStoreResult<TData, TVariables>
    ): Promise<InfinityStoreResult<TData, TVariables>>
}

/**
 * The processed/combined result of a GraphQL query.
 */
export interface InfinityStoreResult<TData = any, TVariables extends AnyVariables = AnyVariables> {
    data?: TData

    /**
     * Set if there was an error fetching the data. When set, no more data will be fetched.
     */
    error?: Error

    /**
     * The set of variables to use for the next fetch. If not set no more data will be fetched.
     */
    nextVariables?: Partial<TVariables>
}

/**
 * The state of the infinity query store.
 */
interface InfinityStoreResultState<TData = any, TVariables extends AnyVariables = AnyVariables>
    extends InfinityStoreResult<TData, TVariables> {
    /**
     * Whether a GraphQL request is currently in flight.
     */
    fetching: boolean
}

// This needs to be exported so that TS type inference can work in SvelteKit generated files.
export interface InfinityQueryStore<TData = any, TVariables extends AnyVariables = AnyVariables, TSnapshot = any>
    extends Readable<InfinityStoreResultState<TData, TVariables>> {
    /**
     * Reruns the query with the next set of query variables.
     *
     * @remarks
     * A new query will only be executed if there is no query currently in flight and {@link InfinityStoreResult.nextVariables}
     * is set.
     */
    fetchMore: () => void

    /**
     * Fetches data while the given predicate is true. Using this function is different f
     * rom calling {@link fetchMore} in a loop, because it will set/unset the fetching state
     * only once.
     */
    fetchWhile: (predicate: (data: TData) => boolean) => Promise<void>

    /**
     * Restores the data state from a snapshot, which is returned by {@link capture}.
     *
     * @param snapshot - The snapshot to restore.
     * @returns A promise that resolves when the data has been restored.
     */
    restore: (snapshot: TSnapshot | undefined) => Promise<void>

    /**
     * Captures the current data state to a snapshot that can be used to restore the data later.
     * @returns The snapshot.
     */
    capture: () => TSnapshot | undefined

    /**
     * Resets the store to its initial state.
     */
    reset: () => void
}

/**
 * Function to create a store to manage "infinity scroll" style queries.
 *
 * @param args - a {@link InfinityQueryArgs} object to pass a `query, `variables` and other options to manage infinity scroll.
 * @returns a {@link InfinityQueryStore} of query results.
 *
 * @remarks
 * `infinityQuery` uses {@link InfinityQueryArgs.client} to execute {@link InfinityQueryArgs.query}
 * with the given {@link InfinityQueryArgs.variables}.
 *
 * The caller can call {@link InfinityQueryStore.fetchMore} to fetch more data. The store will
 * call {@link InfinityQueryArgs.mapResult} to process the query result, combine it with the previous result
 * and to compute the query variables for the next fetch, if any.
 *
 * Calling this function will prefetch the initial data, i.e. the data is fetch before the store is
 * subscribed to.
 */
export function infinityQuery<
    TData = any,
    TPayload = any,
    TVariables extends AnyVariables = AnyVariables,
    TSnapshot = void
>(args: InfinityQueryArgs<TData, TPayload, TVariables, TSnapshot>): InfinityQueryStore<TData, TVariables, TSnapshot> {
    const initialVariables = Promise.resolve(args.variables)

    async function fetch(
        variables: Partial<TVariables>,
        previousResult: InfinityStoreResult<TData, TVariables>
    ): Promise<InfinityStoreResult<TData, TVariables>> {
        const result = args.map(
            await initialVariables.then(initialVariables =>
                args.client.query(args.query, { ...initialVariables, ...variables })
            )
        )
        if (args.merge) {
            result.data = args.merge(previousResult.data, result.data)
        }
        return result
    }

    const initialState: InfinityStoreResultState<TData, TVariables> = { fetching: true }
    const store = writable(initialState)
    const restoreStrategy = args.createRestoreStrategy?.({ store, fetch })

    // Prefetch data. We don't want to wait until the store is subscribed to. That allows us to use this function
    // inside a data loader and the data will be prefetched before the component is rendered.
    fetch({}, {}).then(result => {
        store.update(current => {
            // Only set the initial state if we haven't already started another fetch process,
            // e.g. when restoring the state.
            if (current === initialState) {
                return { ...result, fetching: false }
            }
            return current
        })
    })

    /**
     * Resolves when the store is not fetching anymore.
     */
    function waitTillReady(): Promise<void> {
        let unsubscribe: () => void
        return new Promise<void>(resolve => {
            unsubscribe = store.subscribe(current => {
                if (!current.fetching) {
                    resolve()
                }
            })
        }).finally(() => unsubscribe())
    }

    return {
        subscribe: store.subscribe,
        fetchMore: () => {
            const previous = get(store)

            if (previous.fetching) {
                // When a fetch is already in progress, we don't want to start another one for the same variables.
                return
            }

            if (previous.nextVariables && !previous.error) {
                store.set({ ...previous, fetching: true })
                fetch(previous.nextVariables, previous).then(result => {
                    store.update(current => {
                        if (previous.nextVariables === current.nextVariables) {
                            return { ...result, fetching: false }
                        }
                        return current
                    })
                })
            }
        },
        fetchWhile: async predicate => {
            // We need to wait until the store is not fetching anymore to ensure that we don't start
            // another fetch process while one is already in progress.
            await waitTillReady()
            const current = get(store)

            store.set({ ...current, fetching: true })

            let result: InfinityStoreResult<TData, TVariables> = current
            while (!result.error && result.nextVariables && result.data && predicate(result.data)) {
                result = await fetch(result.nextVariables, result)
            }

            store.set({ ...result, fetching: false })
        },
        capture: () => restoreStrategy?.capture(get(store)),
        restore: snapshot => {
            if (restoreStrategy && snapshot) {
                return restoreStrategy.restore(snapshot)
            }
            return Promise.resolve()
        },
        reset: () => store.set(initialState),
    }
}

/**
 * A restore strategy captures and restores the data state of a query.
 */
interface RestoreStrategy<TSnapshot, TData> {
    capture(result: InfinityStoreResult<TData>): TSnapshot | undefined
    restore(snapshot: TSnapshot): Promise<void>
}

// This needs to be exported so that TS type inference can work in SvelteKit generated files.
export interface IncrementalRestoreStrategySnapshot<TVariables extends AnyVariables> {
    count: number
    variables?: Partial<TVariables>
    nonce: string
}

// We use this to indentify snapshots that were created in the current "session", which
// means there is a high chance that the data is still in the cache.
const NONCE = uniqueID('repeat-restore')

/**
 * The incremental restore strategy captures and restores the data by counting the number of items.
 * It will fetch more data until the count matches the snapshot count.
 *
 * This strategy is useful when every fetch returns a fixed number of items (i.e. after a cursor).
 * In this case we want to make use of our GraphQL client's caching strategy and simply
 * "replay" the previous fetches.
 *
 * This strategy works well when GraphQL requests are cached. To avoid waterfall requests in case the
 * data is not cached, the strategy will fall back to requesting the data once with query variables
 * from the snapshot.
 */
export class IncrementalRestoreStrategy<TData, TVariables extends AnyVariables>
    implements RestoreStrategy<IncrementalRestoreStrategySnapshot<TVariables>, TData>
{
    constructor(
        private api: InfinityAPI<TData, TVariables>,
        /**
         * A function to map the data to a number. This number will be used to count the items.
         */
        private mapper: (data: TData) => number,
        /**
         * A function to map the data to query variables. These variables will be used to fetch the data
         * once when if there is a chance that the data is not in the cache (fallback).
         */
        private variablesMapper?: (data: TData) => Partial<TVariables>
    ) {}

    public capture(result: InfinityStoreResult<TData>): IncrementalRestoreStrategySnapshot<TVariables> | undefined {
        return result.data
            ? {
                  count: this.mapper(result.data),
                  variables: this.variablesMapper ? this.variablesMapper(result.data) : undefined,
                  nonce: NONCE,
              }
            : undefined
    }

    public async restore(snapshot: IncrementalRestoreStrategySnapshot<TVariables>): Promise<void> {
        this.api.store.set({ fetching: true })
        const result = await (snapshot.nonce !== NONCE && snapshot.variables
            ? this.api.fetch(snapshot.variables, {})
            : this.fetch(snapshot))
        this.api.store.set({ ...result, fetching: false })
    }

    private async fetch(
        snapshot: IncrementalRestoreStrategySnapshot<TVariables>
    ): Promise<InfinityStoreResult<TData, TVariables>> {
        let current: InfinityStoreResult<TData, TVariables> = { nextVariables: {} }
        while (current.nextVariables && ((current.data && this.mapper(current.data)) || 0) < snapshot.count) {
            current = await this.api.fetch(current.nextVariables, current)
            if (current.error || !current.data) {
                break
            }
        }
        return current
    }
}

/**
 * A restore strategy that overwrites the current store state with the response of a new query.
 * The strategy uses the query variables form the snapshot to fetch the data.
 */
export class OverwriteRestoreStrategy<TData, TVariables extends AnyVariables>
    implements RestoreStrategy<{ variables: Partial<TVariables> }, TData>
{
    constructor(
        private api: InfinityAPI<TData, TVariables>,
        private variablesMapper: (data: TData) => Partial<TVariables>
    ) {}

    capture(result: InfinityStoreResult<TData, TVariables>): { variables: Partial<TVariables> } | undefined {
        if (!result.data) {
            return undefined
        }
        const variables = this.variablesMapper(result.data)
        return variables ? { variables } : undefined
    }

    async restore(snapshot: { variables: Partial<TVariables> }): Promise<void> {
        this.api.store.set({ fetching: true })
        const result = await this.api.fetch(snapshot.variables, {})
        this.api.store.set({ ...result, fetching: false })
    }
}

/**
 * Converts an OperationResult (urlql) to a GraphQLResult (sourcegraph/http-client).
 */
export function toGraphQLResult<TData = any, TVariables extends AnyVariables = AnyVariables>(
    result: OperationResult<TData, TVariables>
): GraphQLResult<TData> {
    return result.error || !result.data
        ? {
              ...result,
              data: result.data ?? null,
              errors: result.error?.graphQLErrors ?? [],
          }
        : {
              ...result,
              data: result.data,
              errors: undefined,
          }
}

/**
 * Given an {@link OperationResult}, this function returns the mapped data or throws the error.
 * To be used toghether with a promise that resolves to a GraphQL response. This ensures
 * that the promise rejects when the GraphQL response contains an error.
 *
 * @param mapper - A function to map the data from the result.
 */
export function mapOrThrow<T extends OperationResult, U>(mapper: (result: T) => U): (result: T) => U {
    return (result: T) => {
        if (result.error) {
            throw result.error
        }
        return mapper(result)
    }
}
