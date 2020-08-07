import React, { useCallback, useState } from 'react'
import { EMPTY, from, Observable } from 'rxjs'
import { switchMap } from 'rxjs/operators'
import { Toggle } from '../../../shared/src/components/Toggle'
import { PlatformContextProps } from '../../../shared/src/platform/context'
import { SettingsCascadeProps } from '../../../shared/src/settings/settings'
import { eventLogger } from '../tracking/eventLogger'
import { isExtensionAdded } from './extension/extension'
import { useEventObservable } from '../../../shared/src/util/useObservable'

interface Props extends SettingsCascadeProps, PlatformContextProps<'updateSettings'> {
    /** The extension that this element is for. */
    extensionID: string
    enabled: boolean
    className?: string
}

export const ExtensionToggle: React.FunctionComponent<Props> = ({
    settingsCascade,
    platformContext,
    extensionID,
    enabled,
    className,
}) => {
    const [optimisticEnabled, setOptimisticEnabled] = useState(enabled)
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
                        setOptimisticEnabled(enabled)
                        return from(
                            platformContext.updateSettings(highestPrecedenceSubject.subject.id, {
                                path: ['extensions', extensionID],
                                value: enabled,
                            })
                        )
                    })
                ),
            [extensionID, platformContext, settingsCascade]
        )
    )

    const title = optimisticEnabled ? 'Click to disable' : 'Click to enable'

    return (
        <Toggle
            value={optimisticEnabled}
            onToggle={nextToggle}
            title={title}
            className={className}
            dataTest={`extension-toggle-${extensionID}`}
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
