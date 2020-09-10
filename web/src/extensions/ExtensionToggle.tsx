import React, { useCallback, useState } from 'react'
import { EMPTY, from, Observable } from 'rxjs'
import { switchMap } from 'rxjs/operators'
import { Toggle } from '../../../shared/src/components/Toggle'
import { ToggleBig } from '../../../shared/src/components/ToggleBig'
import { PlatformContextProps } from '../../../shared/src/platform/context'
import { SettingsCascadeProps } from '../../../shared/src/settings/settings'
import { eventLogger } from '../tracking/eventLogger'
import { isExtensionAdded } from './extension/extension'
import { useEventObservable } from '../../../shared/src/util/useObservable'

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

type ExtensionToggleState = 'enabled' | 'disabled' | 'askingForPermission'

type ExtensionToggleAction =
    | {
          type: 'enable'
          isExtensionAdded: boolean
      }
    | { type: 'disable' }
    | { type: 'permissionGiven' }
    | { type: 'permissionDenied' }

function extensionToggleReducer(state: ExtensionToggleState, action: ExtensionToggleAction): ExtensionToggleState {
    switch (state) {
        case 'enabled':
            if (action.type === 'disable') {
                return 'disabled'
            }

        case 'disabled':
            if (action.type === 'enable') {
                return action.isExtensionAdded ? 'enabled' : 'askingForPermission'
            }

        case 'askingForPermission':
            if (action.type === 'permissionGiven') {
                return 'enabled'
            }

            if (action.type === 'permissionDenied') {
                return 'disabled'
            }
    }
    // state is unchanged for unexpected actions
    return state
}

/**
 * TODO: Refactor to using reducer bc can no longer block
 * thread with dialog (new modal)
 *
 * might not be able to use effect + reducer combo here,
 * but take the gist of state/flow from reducer
 */

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
    const [askingForPermission, setAskingForPermission] = useState<{ enabled: boolean } | false>(false)

    const [nextToggle] = useEventObservable(
        useCallback(
            (toggles: Observable<boolean>) =>
                toggles.pipe(
                    switchMap(enabled => {
                        if (settingsCascade.subjects === null) {
                            return EMPTY
                        }

                        // Only operate on the highest precedence settings, for simplicity.
                        const subjects = settingsCascade.subjects
                        if (subjects.length === 0) {
                            return EMPTY
                        }
                        const highestPrecedenceSubject = subjects[subjects.length - 1]
                        if (!highestPrecedenceSubject || !highestPrecedenceSubject.subject.viewerCanAdminister) {
                            return EMPTY
                        }

                        if (
                            !isExtensionAdded(settingsCascade.final, extensionID) &&
                            !confirmAddExtension(extensionID)
                        ) {
                            return EMPTY
                        }

                        eventLogger.log('ExtensionToggled', { extension_id: extensionID })

                        if (onToggleChange) {
                            onToggleChange(enabled)
                        }
                        setOptimisticEnabled(enabled)
                        return from(
                            platformContext.updateSettings(highestPrecedenceSubject.subject.id, {
                                path: ['extensions', extensionID],
                                value: enabled,
                            })
                        )
                    })
                ),
            [extensionID, platformContext, settingsCascade, onToggleChange]
        )
    )

    // core logic
    function toggle(enabled: boolean) {
        eventLogger.log('ExtensionToggled', { extension_id: extensionID })

        if (onToggleChange) {
            onToggleChange(enabled)
        }
        setOptimisticEnabled(enabled)
    }

    const denyPermission = useCallback(() => {
        // noop
        setOptimisticEnabled(false)
        setAskingForPermission(false)
    }, [])

    const givePermission = useCallback(() => {
        // noop
        setOptimisticEnabled(true)
        setAskingForPermission(false)
    }, [])

    const title = optimisticEnabled ? 'Click to disable' : 'Click to enable'

    const props = {}

    return big ? (
        <ToggleBig
            value={optimisticEnabled}
            onToggle={nextToggle}
            title={title}
            className={className}
            dataTest={`extension-toggle-${extensionID}`}
            disabled={userCannotToggle}
            onHover={onHover}
        />
    ) : (
        <Toggle
            value={optimisticEnabled}
            onToggle={nextToggle}
            title={title}
            className={className}
            dataTest={`extension-toggle-${extensionID}`}
            disabled={userCannotToggle}
        />
    )
}

/**
 * Shows a modal confirmation prompt to the user confirming whether to add an extension.
 */
function confirmAddExtension(extensionID: string): boolean {
    return confirm(
        `Add Sourcegraph extension ${extensionID}?\n\nIt can:\n- Read repositories and files you view using Sourcegraph\n- Read and change your Sourcegraph settings`
    )
}
