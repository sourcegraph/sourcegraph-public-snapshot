import * as React from 'react'
import { useCallback, useEffect, useState } from 'react'

import { SettingsCascadeOrError } from 'out/src/settings/settings'
import { Subscribable, Unsubscribable } from 'rxjs'

import { SimpleActionItem } from '../actions/SimpleActionItem'

export interface OpenInEditorActionItemProps {
    settingsSubscribable: Subscribable<SettingsCascadeOrError>
}

export const OpenInEditorActionItem: React.FunctionComponent<OpenInEditorActionItemProps> = ({
    settingsSubscribable,
}) => {
    const [settingsCascadeOrError, setSettingsCascadeOrError] = useState<SettingsCascadeOrError | undefined>(undefined)
    const [settingSubscription, setSettingSubscription] = useState<Unsubscribable | null>(null)
    const settings =
        settingsCascadeOrError?.final && !('message' in settingsCascadeOrError.final) // isErrorLike fails with some TypeScript error
            ? settingsCascadeOrError.final
            : undefined
    useEffect(() => {
        setSettingSubscription(
            settingsSubscribable.subscribe(settings => (settings.final ? setSettingsCascadeOrError(settings) : null))
        )

        return () => {
            settingSubscription?.unsubscribe()
        }
    }, [settingSubscription, settingsSubscribable])

    const onClick = useCallback(() => {
        if (settings?.experimentalFeatures) {
            alert('test')
        }
    }, [settings])

    return (
        <SimpleActionItem
            tooltip={settings ? settings['openInEditor.editorId'] : 'No data'}
            className={settings ? settings['openInEditor.editorId'] : 'disabled'}
            iconURL=""
            onClick={onClick}
        />
    )
}
