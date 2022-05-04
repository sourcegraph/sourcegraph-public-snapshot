import React, { createContext, useEffect } from 'react'

import { migrateLocalStorageToTemporarySettings } from './migrateLocalStorageToTemporarySettings'
import { TemporarySettingsStorage } from './TemporarySettingsStorage'

export const TemporarySettingsContext = createContext<TemporarySettingsStorage>(
    new TemporarySettingsStorage(null, false)
)
TemporarySettingsContext.displayName = 'TemporarySettingsContext'

/**
 * React context provider for the temporary settings.
 * The web app needs to be wrapped around this.
 */
export const TemporarySettingsProvider: React.FunctionComponent<
    React.PropsWithChildren<{
        temporarySettingsStorage: TemporarySettingsStorage
    }>
> = ({ children, temporarySettingsStorage }) => {
    // On first run, migrate the settings from the local storage to the temporary storage.
    useEffect(() => {
        const migrate = async (): Promise<void> => {
            await migrateLocalStorageToTemporarySettings(temporarySettingsStorage)
        }

        migrate().catch(console.error)
    }, [temporarySettingsStorage])

    // Dispose temporary settings storage on unmount.
    useEffect(() => () => temporarySettingsStorage.dispose(), [temporarySettingsStorage])

    return (
        <TemporarySettingsContext.Provider value={temporarySettingsStorage}>
            {children}
        </TemporarySettingsContext.Provider>
    )
}
