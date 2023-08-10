import { useEffect, useState } from 'react'

import { getVersion } from '@tauri-apps/api/app'
import { TauriEvent, listen } from '@tauri-apps/api/event'
import {
    type UpdateManifest,
    type UpdateStatus,
    type UpdateStatusResult,
    checkUpdate,
    installUpdate,
} from '@tauri-apps/api/updater'
import { SemVer } from 'semver'

const StartUpdateCheckDelayMs = 3 * 1000 // time to wait to start check
const StartInstallCheckDelayMs = 3 * 1000 // time to wait to start installing
const UpdateCheckIntervalMs = 10 * 60 * 1000 // time in betwen checks

type Stage = 'IDLE' | 'CHECKING' | 'INSTALLING' | UpdateStatus

export interface UpdateInfo {
    stage: Stage
    hasNewVersion: boolean
    version?: string
    newVersion?: string
    description?: string
    startInstall?: () => void
    checkNow?: (force: boolean) => void
    error?: string
}

export interface UpdaterSettings {
    keepChecking?: boolean
}

/**
 * A React hook to check for app updates.
 *
 * @returns An object with:
 * - `stage`: The current stage of the update check ('IDLE' | 'CHECKING' | 'INSTALLING' | UpdateStatus)
 * - `hasNewVersion`: Whether there is an available update
 * - `version`: The current app version
 * - `newVersion`: The available update version
 * - `description`: The update description (supports markdown)
 * - `startInstall()`: A function to start installing the available update
 * - `checkNow(force)`: A function to manually check for updates. Pass `true` to force a check even if an update was recently found.
 * - `error`: Any error that occurred during update checking
 */
export function useUpdater({ keepChecking }: UpdaterSettings = { keepChecking: true }): UpdateInfo {
    const [lastCheck, setLastCheck] = useState<UpdateInfo>({
        stage: 'CHECKING',
        hasNewVersion: false,
        startInstall: () => {
            setLastCheck((check: UpdateInfo): UpdateInfo => ({ ...check, stage: 'INSTALLING', error: undefined }))
            setTimeout(() => {
                installUpdate().finally(() => {})
            }, StartInstallCheckDelayMs)
        },
        checkNow: (force: boolean) => {
            setLastCheck((check: UpdateInfo): UpdateInfo => {
                if (check.stage === 'CHECKING') {
                    return check
                }
                if (check.hasNewVersion && !force) {
                    return check
                }
                startUpdateCheck()
                return { ...check, stage: 'CHECKING', error: undefined }
            })
        },
    })

    function startUpdateCheck(): void {
        setTimeout(() => {
            checkUpdate()
                .then(result => {
                    setLastCheck((check: UpdateInfo): UpdateInfo => {
                        const version = result?.manifest?.version || 'unknown'
                        return {
                            ...check,
                            stage: 'IDLE',
                            hasNewVersion: result.shouldUpdate,
                            newVersion: version,
                            error: undefined,
                        }
                    })
                })
                .catch(() => {})
        }, StartUpdateCheckDelayMs)
    }

    useEffect(() => {
        startUpdateCheck()
    }, [])

    useEffect(() => {
        if (keepChecking) {
            const timer = setInterval(() => {
                if (lastCheck.stage !== 'IDLE') {
                    return
                }
                lastCheck.checkNow?.(false)
            }, UpdateCheckIntervalMs)
            return () => clearInterval(timer)
        }
        return () => {}
    }, [keepChecking, lastCheck])

    useEffect(() => {
        const unregisterAvailable = listen<UpdateManifest>(TauriEvent.UPDATE_AVAILABLE, ({ payload }) => {
            setLastCheck(
                (check: UpdateInfo): UpdateInfo => ({
                    ...check,
                    stage: 'IDLE',
                    hasNewVersion: new SemVer(payload.version).compare(check.version || '0.0.0') === 1,
                    newVersion: payload.version,
                    description: payload.body,
                })
            )
        })
        const unlistenUpdate = listen<UpdateStatusResult>(TauriEvent.STATUS_UPDATE, ({ payload }) => {
            setLastCheck(
                (check: UpdateInfo): UpdateInfo => ({
                    ...check,
                    stage: payload.status,
                    error: payload.error,
                })
            )
        })
        return () => {
            unregisterAvailable.then(callFn => callFn()).catch(() => {})
            unlistenUpdate.then(callFn => callFn()).catch(() => {})
        }
    }, [])

    useEffect(() => {
        getVersion()
            .then(version => setLastCheck((check: UpdateInfo) => ({ ...check, version })))
            .catch(() => {})
    }, [])

    return lastCheck
}
