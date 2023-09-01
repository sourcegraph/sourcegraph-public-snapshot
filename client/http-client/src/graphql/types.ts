/**
 * Shared global Apollo context is used in hooks to specify some context properties
 * and read in apollo links to turn on/off some internal apollo request logic.
 */
export interface ApolloContext {
    /**
     * Turns on/off concurrent/parallel requests apollo link.
     * See `./links/concurrent-requests-link.ts` for more details.
     */
    concurrentRequests?: {
        /** Group requests by this key and run them concurrently/in parallel */
        key?: string

        /** The size of parallel requests queue for particular group of requests */
        limit?: number
    }
}
