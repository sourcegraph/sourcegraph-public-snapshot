/**
 * A store of individual (constant) global values.
 */

import create from 'zustand'

interface Context {
    isSourcegraphDotCom: boolean
}

export const useContext = create<Context>(() => ({
    isSourcegraphDotCom: false,
}))
