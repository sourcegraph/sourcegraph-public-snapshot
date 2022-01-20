import React from 'react'

import { WildcardThemeContext } from '@sourcegraph/wildcard'

export const WildcardThemeProvider: React.FunctionComponent = ({ children }) => (
    <WildcardThemeContext.Provider value={{ isBranded: false }}>{children}</WildcardThemeContext.Provider>
)
