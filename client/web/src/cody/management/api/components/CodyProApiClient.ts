import { createContext } from 'react'

import { type Caller, CodyProApiCaller } from '../client'

export interface CodyProApiClient {
    caller: Caller
}

export const defaultCodyProApiClientContext: { caller: Caller } = { caller: new CodyProApiCaller() }

// Context for supplying a Cody Pro API client to a React component tree.
//
// The default value will be a functional API client that makes HTTP requests
// to the current Sourcegraph instance's backend.
export const CodyProApiClientContext = createContext<CodyProApiClient>(defaultCodyProApiClientContext)
