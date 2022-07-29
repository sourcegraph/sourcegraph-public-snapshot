import React from 'react'

export interface TreeRootContext {
    rootTreeUrl: string
}

export const TreeRootContext = React.createContext<TreeRootContext>({
    rootTreeUrl: '',
})
