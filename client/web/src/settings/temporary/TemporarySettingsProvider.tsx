import { useApolloClient } from '@apollo/client'
import React, { createContext, useEffect, useState } from 'react'

import { AuthenticatedUser } from '../../auth'
import { client } from '../../backend/graphql'

import { TemporarySettingsStorage } from './TemporarySettingsStorage'

export const TemporarySettingsContext = createContext<TemporarySettingsStorage>(
    new TemporarySettingsStorage(client, null)
)
TemporarySettingsContext.displayName = 'TemporarySettingsContext'

/**
 * React context provider for the temporary settings.
 * The web app needs to be wrapped around this.
 */
export const TemporarySettingsProvider: React.FunctionComponent<{
    authenticatedUser: AuthenticatedUser | null
}> = ({ children, authenticatedUser }) => {
    const apolloClient = useApolloClient()

    const [temporarySettingsStorage] = useState<TemporarySettingsStorage>(
        () => new TemporarySettingsStorage(apolloClient, authenticatedUser)
    )

    useEffect(() => () => temporarySettingsStorage.dispose(), [temporarySettingsStorage])

    useEffect(() => {
        temporarySettingsStorage?.setAuthenticatedUser(authenticatedUser)
    }, [temporarySettingsStorage, authenticatedUser])

    return (
        <TemporarySettingsContext.Provider value={temporarySettingsStorage}>
            {children}
        </TemporarySettingsContext.Provider>
    )
}
