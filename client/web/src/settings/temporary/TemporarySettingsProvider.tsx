import { useApolloClient } from '@apollo/client'
import React, { createContext, useEffect, useState } from 'react'

import { TemporarySettingsStorage } from './TemporarySettingsStorage'

export const TemporarySettingsContext = createContext<TemporarySettingsStorage>(
    new TemporarySettingsStorage(null, false)
)
TemporarySettingsContext.displayName = 'TemporarySettingsContext'

/**
 * React context provider for the temporary settings.
 * The web app needs to be wrapped around this.
 */
export const TemporarySettingsProvider: React.FunctionComponent<{
    isAuthenticatedUser: boolean
}> = ({ children, isAuthenticatedUser }) => {
    const apolloClient = useApolloClient()

    const [temporarySettingsStorage] = useState<TemporarySettingsStorage>(
        () => new TemporarySettingsStorage(apolloClient, isAuthenticatedUser)
    )

    useEffect(() => () => temporarySettingsStorage.dispose(), [temporarySettingsStorage])

    return (
        <TemporarySettingsContext.Provider value={temporarySettingsStorage}>
            {children}
        </TemporarySettingsContext.Provider>
    )
}
