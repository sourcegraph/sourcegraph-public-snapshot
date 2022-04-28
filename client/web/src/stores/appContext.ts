import create from 'zustand'

import { AuthenticatedUser } from '@sourcegraph/shared/src/auth'

/**
 * A store of individual (constant) global values.
 */

import { createSingle } from './utils'

interface AppContext {
    isSourcegraphDotCom: boolean
    authenticatedUser: AuthenticatedUser | null
}

export const useAppContext = create<AppContext>(() => ({
    isSourcegraphDotCom: false,
    authenticatedUser: null,
}))

export const useIsSourcegraphDotCom = createSingle(useAppContext, state => state.isSourcegraphDotCom)
export const useAuthenticatedUser = createSingle(useAppContext, state => state.authenticatedUser)
