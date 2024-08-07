import { writable, type Readable } from 'svelte/store'

import { type GraphQLClient } from '$lib/graphql'

import { CreateSearchJob, ValidateSearchJob } from './searchJob.gql'

interface Store {
    validating: boolean
    validationError: Error | null
    creating: boolean
    creationError: Error | null
}

export class SearchJob implements Readable<Store> {
    private store = writable<Store>({ validating: false, validationError: null, creating: false, creationError: null })

    constructor(private client: GraphQLClient, private query: string) {
        this.validate().catch(() => {
            // intentionally ignored
        })
    }

    /**
     * Validates the query and returns an error if it is invalid.
     */
    private async validate(): Promise<void> {
        this.store.update($store => ({ ...$store, validating: true, validationError: null }))
        try {
            const response = await this.client
                .query(ValidateSearchJob, { query: this.query }, { requestPolicy: 'network-only' })
                .toPromise()
            if (response.error) {
                throw response.error
            }
        } catch (error) {
            this.store.update($store => ({ ...$store, validationError: error as Error }))
            throw error
        } finally {
            this.store.update($store => ({ ...$store, validating: false }))
        }
    }

    /**
     * Creates a new search job.
     */
    public async create(): Promise<void> {
        this.store.update($store => ({ ...$store, creating: true, creationError: null }))
        try {
            const response = await this.client.mutation(CreateSearchJob, { query: this.query }).toPromise()
            if (response.error) {
                throw response.error
            }
        } catch (error) {
            this.store.update($store => ({ ...$store, creationError: error as Error }))
            throw error
        } finally {
            this.store.update($store => ({ ...$store, creating: false }))
        }
    }

    public subscribe(run: (value: Store) => void) {
        return this.store.subscribe(run)
    }
}
