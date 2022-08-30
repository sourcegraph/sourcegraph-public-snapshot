import * as React from 'react'
import { useCallback, useEffect, useMemo, useState } from 'react'

import { SettingsCascadeOrError } from 'out/src/settings/settings'
import { Unsubscribable } from 'rxjs'

import { SimpleActionItem } from '../actions/SimpleActionItem'
import { PlatformContext } from '../platform/context'

import { buildEditorUrl } from './build-url'
import { getEditor } from './editors'

export interface OpenInEditorActionItemProps {
    platformContext: PlatformContext
    assetsRoot?: string
}

export const OpenInEditorActionItem: React.FunctionComponent<OpenInEditorActionItemProps> = props => {
    const [settingsCascadeOrError, setSettingsCascadeOrError] = useState<SettingsCascadeOrError | undefined>(undefined)
    const [settingSubscription, setSettingSubscription] = useState<Unsubscribable | null>(null)
    const settings =
        settingsCascadeOrError?.final && !('message' in settingsCascadeOrError.final) // isErrorLike fails with some TypeScript error
            ? settingsCascadeOrError.final
            : undefined
    const editorUrl = useMemo(() => {
        if (settings) {
            try {
                return buildEditorUrl(
                    undefined, // TODO: Add ViewComponent
                    settings.openInEditor,
                    props.platformContext.sourcegraphURL
                )
            } catch {
                // TODO: Swallowing errors this way is not nice
                return undefined
            }
        }

        return undefined
    }, [props.platformContext.sourcegraphURL, settings])

    const assetsRoot = props.assetsRoot ?? (window.context?.assetsRoot || '')
    const editor = editorUrl ? getEditor(settings?.openInEditor?.editorId || '') : undefined

    useEffect(() => {
        setSettingSubscription(
            props.platformContext.settings.subscribe(settings =>
                settings.final ? setSettingsCascadeOrError(settings) : null
            )
        )

        return () => {
            settingSubscription?.unsubscribe()
        }
    }, [settingSubscription, props.platformContext.settings])

    const onClick = useCallback(() => {
        if (editorUrl) {
            alert(`Opening ${editorUrl}`);
        } else {
            alert('Opening setup popover')
        }
    }, [editorUrl])

    return (
        <SimpleActionItem
            tooltip={editorUrl ? `Open file in ${editor?.name}` : 'Set your preferred editor'}
            className="enabled"
            iconURL={editor ? `${assetsRoot}/img/editors/jetbrains.svg` : `${assetsRoot}/img/open-in-editor.svg`}
            onClick={onClick}
        />
    )
}
