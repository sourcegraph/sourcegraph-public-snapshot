import React from 'react'

export interface IOptionsPageContext {
    // Repo blocklist
    blocklist?: {
        enabled: boolean
        content: string
    }
    onBlocklistChange: (enabled: boolean, content: string) => void

    // Option flags
    optionFlags: { key: string; label: string; value: boolean }[]
    onChangeOptionFlag: (key: string, value: boolean) => void
}

export const OptionsPageContext = React.createContext<IOptionsPageContext>({} as IOptionsPageContext)
