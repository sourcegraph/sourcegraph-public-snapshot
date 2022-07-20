import React from 'react'

export interface TreeContext {
    rootTreeUrl: string
}

export const TreeContext = React.createContext<TreeContext>({
    rootTreeUrl: '',
})
