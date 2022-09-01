import * as React from 'react'
import { useCallback, useEffect, useState } from 'react'

import { Unsubscribable } from 'rxjs'

import { PlatformContext } from '@sourcegraph/shared/out/src/platform/context'
import { SettingsCascadeOrError } from '@sourcegraph/shared/out/src/settings/settings'

import { SimpleActionItem } from '../../../shared/src/actions/SimpleActionItem'

import { getEditorSettingsErrorMessage } from './build-url'
import type { EditorSettings } from './editor-settings'
import { getEditor } from './editors'
import { useOpenCurrentUrlInEditor } from './useOpenCurrentUrlInEditor'

export interface OpenInEditorActionItemProps {
    platformContext: PlatformContext
    assetsRoot?: string
}

export const OpenInEditorActionItem: React.FunctionComponent<OpenInEditorActionItemProps> = props => {
    const assetsRoot = props.assetsRoot ?? (window.context?.assetsRoot || '')

    const [settingsCascadeOrError, setSettingsCascadeOrError] = useState<SettingsCascadeOrError | undefined>(undefined)
    const settings =
        settingsCascadeOrError?.final && !('message' in settingsCascadeOrError.final) // isErrorLike fails with some TypeScript error
            ? settingsCascadeOrError.final
            : undefined
    const [settingSubscription, setSettingSubscription] = useState<Unsubscribable | null>(null)

    const openCurrentUrlInEditor = useOpenCurrentUrlInEditor()

    const editorSettingsErrorMessage = getEditorSettingsErrorMessage(
        settings?.openInEditor,
        props.platformContext.sourcegraphURL
    )
    const editor = !editorSettingsErrorMessage
        ? getEditor((settings?.openInEditor as EditorSettings | undefined)?.editorId || '')
        : undefined

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
        if (editor) {
            openCurrentUrlInEditor(settings?.openInEditor, props.platformContext.sourcegraphURL)
        } else {
            alert('Opening setup popover.')
        }
    }, [editor, openCurrentUrlInEditor, props.platformContext.sourcegraphURL, settings?.openInEditor])

    return (
        <SimpleActionItem
            tooltip={editor ? `Open file in ${editor?.name}` : 'Set your preferred editor'}
            className="enabled"
            iconURL={editor ? `${assetsRoot}/img/editors/${editor.id}.svg` : `${assetsRoot}/img/open-in-editor.svg`}
            onClick={onClick}
        />
    )
}
