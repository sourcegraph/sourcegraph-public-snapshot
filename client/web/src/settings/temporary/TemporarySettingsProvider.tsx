import React, { createContext, useEffect } from 'react'

import { AuthenticatedUser } from '../../auth'

import { TemporarySettingsStorage } from './TemporarySettingsStorage'

export const TemporarySettingsContext = createContext<TemporarySettingsStorage>(new TemporarySettingsStorage())

export const TemporarySettingsProvider: React.FunctionComponent<{ authenticatedUser: AuthenticatedUser | null }> = ({
    children,
    authenticatedUser,
}) => {
    const temporarySettings = React.useRef(new TemporarySettingsStorage())

    useEffect(() => {
        temporarySettings.current.setAuthenticatedUser(authenticatedUser)
    }, [authenticatedUser])

    return (
        <TemporarySettingsContext.Provider value={temporarySettings.current}>
            {children}
        </TemporarySettingsContext.Provider>
    )
}
