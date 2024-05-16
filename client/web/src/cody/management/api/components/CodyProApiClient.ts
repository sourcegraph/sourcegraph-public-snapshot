import { createContext } from 'react'

import { Client, Caller, CodyProApiCaller } from '../client'

export interface CodyProApiClient {
    client: Client
    caller: Caller
}

// Context for supplying a Cody Pro API client to a React component tree.
//
// The default value will be a functional API client that makes HTTP requests
// to the current Sourcegraph instance's backend.
export const CodyProApiClientContext = createContext<CodyProApiClient>({
    client: new Client(),
    caller: new CodyProApiCaller(),
})
