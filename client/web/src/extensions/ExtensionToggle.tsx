import React, { useCallback, useMemo, useState } from 'react'

import { Observable, of } from 'rxjs'
import { catchError, map, switchMap } from 'rxjs/operators'

import { Toggle } from '@sourcegraph/branded/src/components/Toggle'
import { ToggleBig } from '@sourcegraph/branded/src/components/ToggleBig'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { SettingsCascadeProps, SettingsSubject } from '@sourcegraph/shared/src/settings/settings'
import { useEventObservable } from '@sourcegraph/wildcard'

import { eventLogger } from '../tracking/eventLogger'

import { isExtensionAdded } from './extension/extension'
import { ExtensionPermissionModal } from './ExtensionPermissionModal'

interface Props extends SettingsCascadeProps, PlatformContextProps<'updateSettings'> {
    /** The id of the extension that this element is for. */
    extensionID: string
    enabled: boolean
    className?: string
    /** Render big toggle */
    big?: boolean
    userCannotToggle?: boolean
    onHover?: (value: boolean) => void
    /** Additional logic to run on toggle */
    onToggleChange?: (enabled: boolean) => void
    /** Additional logic to run on update error */
    onToggleError?: (revertedValue: OptimisticUpdateFailure<boolean>) => void
    /** Settings subject that the toggle should act upon */
    subject: SettingsSubject | null

    /** Render text left of the toggle based on optimistic enablement state */
    renderText?: (enabled: boolean) => JSX.Element
}

export interface OptimisticUpdateFailure<T> {
    previousValue: T
    optimisticValue: T
    error: Error
}

/**
 * Creates a pipeline to use with our `useEventObservable` hook.
 * Helps with error handling for optimistic updates.
 *
 * How it works:
 * - Wraps the optimistic update request promise with another
 * promise that only resolves if the inner promise is rejected.
 * The resulting observable will then only emit on error.
 * - Cancels subscriptions to old promises by way of `switchMap`; we
 * only care about the latest optimistic update
 *
 * TODO: Make it work with observables as well.
 *
 * @param onError Function called with the previous value and the optimistic value, in case you
 * want to display the optimistic value in an error message.
 */
function createOptimisticRollbackPipeline<T>(
    onError: (optimisticUpdateFailure: OptimisticUpdateFailure<T>) => void
): (
    optimisticUpdates: Observable<{ previousValue: T; optimisticValue: T; promise: Promise<void> }>
) => Observable<void> {
    return function optimisticRollbackPipeline(optimisticUpdates) {
        return optimisticUpdates.pipe(
            switchMap(
                ({ promise, previousValue, optimisticValue }) =>
                    new Promise<OptimisticUpdateFailure<T>>(resolve => {
                        promise.catch((error: Error) => resolve({ error, previousValue, optimisticValue }))
                    })
            ),
            map(optimisticUpdateFailure => {
                onError(optimisticUpdateFailure)
            }),
            catchError<void, Observable<void>>(() => of())
        )
    }
}

export const ExtensionToggle: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
    settingsCascade,
    platformContext,
    extensionID,
    enabled,
    className,
    big,
    userCannotToggle,
    onHover,
    onToggleChange,
    onToggleError,
    subject,
    renderText,
}) => {
    const [optimisticEnabled, setOptimisticEnabled] = useState(enabled)
    const [askingForPermission, setAskingForPermission] = useState<boolean>(false)

    const onOptimisticError = useCallback(
        (optimisticUpdateFailure: OptimisticUpdateFailure<boolean>) => {
            setOptimisticEnabled(optimisticUpdateFailure.previousValue)
            onToggleError?.(optimisticUpdateFailure)
        },
        [onToggleError, setOptimisticEnabled]
    )

    const [nextOptimisticUpdate] = useEventObservable<
        { previousValue: boolean; optimisticValue: boolean; promise: Promise<void> },
        void
    >(useMemo(() => createOptimisticRollbackPipeline(onOptimisticError), [onOptimisticError]))

    const updateEnablement = useCallback(
        (enabled: boolean) => {
            if (!subject) {
                return
            }

            eventLogger.log(
                'ExtensionToggled',
                { extension_id: extensionID, subject_type: subject.__typename },
                { extension_id: extensionID, subject_type: subject.__typename }
            )

            if (onToggleChange) {
                onToggleChange(enabled)
            }
            setOptimisticEnabled(enabled)

            nextOptimisticUpdate({
                previousValue: !enabled,
                optimisticValue: enabled,
                promise: platformContext.updateSettings(subject.id, {
                    path: ['extensions', extensionID],
                    value: enabled,
                }),
            })
        },
        [platformContext, extensionID, onToggleChange, nextOptimisticUpdate, subject]
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
        title: userCannotToggle || !subject ? undefined : optimisticEnabled ? 'Click to disable' : 'Click to enable',
        ['data-testid']: `extension-toggle-${extensionID}`,
        disabled: userCannotToggle || !subject,
    }

    return (
        <>
            {renderText?.(optimisticEnabled)}
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
