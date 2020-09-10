import React, { useCallback, useState } from 'react'
import { Toggle } from '../../../shared/src/components/Toggle'
import { ToggleBig } from '../../../shared/src/components/ToggleBig'
import { PlatformContextProps } from '../../../shared/src/platform/context'
import {
    SettingsCascadeProps,
    SettingsCascadeOrError,
    ConfiguredSubjectOrError,
} from '../../../shared/src/settings/settings'
import { eventLogger } from '../tracking/eventLogger'
import { isExtensionAdded } from './extension/extension'
import { ExtensionPermissionModal } from './ExtensionPermissionModal'

interface Props extends SettingsCascadeProps, PlatformContextProps<'updateSettings'> {
    /** The id of the extension that this element is for. */
    extensionID: string
    enabled: boolean
    className?: string
    /** Additional logic to run on toggle */
    onToggleChange?: (enabled: boolean) => void
    /** Render big toggle */
    big?: boolean
    userCannotToggle?: boolean
    onHover?: (value: boolean) => void
}

export const ExtensionToggle: React.FunctionComponent<Props> = ({
    settingsCascade,
    platformContext,
    extensionID,
    enabled,
    className,
    onToggleChange,
    big,
    userCannotToggle,
    onHover,
}) => {
    const [optimisticEnabled, setOptimisticEnabled] = useState(enabled)
    const [askingForPermission, setAskingForPermission] = useState<boolean>(false)

    const updateEnablement = useCallback(
        (enabled: boolean) => {
            const highestPrecedenceSubject = getHighestPrecedenceSubject(settingsCascade)

            if (!highestPrecedenceSubject) {
                return
            }

            eventLogger.log('ExtensionToggled', { extension_id: extensionID })

            if (onToggleChange) {
                onToggleChange(enabled)
            }
            setOptimisticEnabled(enabled)

            // eslint-disable-next-line no-void
            void platformContext.updateSettings(highestPrecedenceSubject.subject.id, {
                path: ['extensions', extensionID],
                value: enabled,
            })
        },
        [platformContext, extensionID, onToggleChange, settingsCascade]
    )

    const onToggle = useCallback(
        (enabled: boolean) => {
            if (!enabled) {
                updateEnablement(false)
            } else if (!isExtensionAdded(settingsCascade.final, extensionID)) {
                setAskingForPermission(true)
            } else {
                updateEnablement(true)
            }
        },
        [updateEnablement, extensionID, settingsCascade]
    )

    const denyPermission = useCallback(() => {
        setAskingForPermission(false)
    }, [])

    const givePermission = useCallback(() => {
        updateEnablement(true)
        setAskingForPermission(false)
    }, [updateEnablement])

    const props: React.ComponentProps<typeof ToggleBig> = {
        onToggle,
        onHover,
        className,
        value: optimisticEnabled,
        title: userCannotToggle ? undefined : optimisticEnabled ? 'Click to disable' : 'Click to enable',
        dataTest: `extension-toggle-${extensionID}`,
        disabled: userCannotToggle,
    }

    return (
        <>
            {big ? <ToggleBig {...props} /> : <Toggle {...props} />}
            {askingForPermission && (
                <ExtensionPermissionModal
                    extensionID={extensionID}
                    givePermission={givePermission}
                    denyPermission={denyPermission}
                />
            )}
        </>
    )
}

/** If this function returns undefined, do not update extension enablement */
function getHighestPrecedenceSubject(settingsCascade: SettingsCascadeOrError): ConfiguredSubjectOrError | undefined {
    if (settingsCascade.subjects === null) {
        return
    }

    // Only operate on the highest precedence settings, for simplicity.
    const subjects = settingsCascade.subjects
    if (subjects.length === 0) {
        return
    }
    const highestPrecedenceSubject = subjects[subjects.length - 1]
    if (!highestPrecedenceSubject || !highestPrecedenceSubject.subject.viewerCanAdminister) {
        return
    }

    return highestPrecedenceSubject
}
