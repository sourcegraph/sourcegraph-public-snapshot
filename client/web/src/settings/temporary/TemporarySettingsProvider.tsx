import React, { createContext, useEffect, useState } from 'react'

import { AuthenticatedUser } from '../../auth'

import { TemporarySettingsStorage } from './TemporarySettingsStorage'

export const TemporarySettingsContext = createContext<TemporarySettingsStorage>(new TemporarySettingsStorage())
TemporarySettingsContext.displayName = 'TemporarySettingsContext'

/**
 * React context provider for the temporary settings.
 * The web app needs to be wrapped around this.
 */
export const TemporarySettingsProvider: React.FunctionComponent<{ authenticatedUser: AuthenticatedUser | null }> = ({
    children,
    authenticatedUser,
}) => {
    const [temporarySettingsStorage] = useState<TemporarySettingsStorage>(() => {
        const storage = new TemporarySettingsStorage()
        storage.setAuthenticatedUser(authenticatedUser)
        return storage
    })

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
