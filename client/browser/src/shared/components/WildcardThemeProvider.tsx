import React from 'react'

import { type WildcardTheme, WildcardThemeContext } from '@sourcegraph/wildcard'

export const WildcardThemeProvider: React.FunctionComponent<React.PropsWithChildren<WildcardTheme>> = ({
    children,
    ...props
}) => <WildcardThemeContext.Provider value={props}>{children}</WildcardThemeContext.Provider>
