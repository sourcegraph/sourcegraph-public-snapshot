import * as React from 'react'
import { useCallback, useEffect, useMemo, useState } from 'react'

import { useLocation } from 'react-router'
import { Unsubscribable } from 'rxjs'

import { PlatformContext } from '@sourcegraph/shared/out/src/platform/context'
import { SettingsCascadeOrError } from '@sourcegraph/shared/out/src/settings/settings'

import { SimpleActionItem } from '../../../shared/src/actions/SimpleActionItem'
import { parseBrowserRepoURL } from '../util/url'

import { buildEditorUrl } from './build-url'
import type { EditorSettings } from './editor-settings'
import { getEditor } from './editors'

export interface OpenInEditorActionItemProps {
    platformContext: PlatformContext
    assetsRoot?: string
}

export const OpenInEditorActionItem: React.FunctionComponent<OpenInEditorActionItemProps> = props => {
    const [settingsCascadeOrError, setSettingsCascadeOrError] = useState<SettingsCascadeOrError | undefined>(undefined)
    const [settingSubscription, setSettingSubscription] = useState<Unsubscribable | null>(null)
    const location = useLocation()
    const { repoName, filePath, range } = parseBrowserRepoURL(location.pathname)
    const settings =
        settingsCascadeOrError?.final && !('message' in settingsCascadeOrError.final) // isErrorLike fails with some TypeScript error
            ? settingsCascadeOrError.final
            : undefined
    const editorUrl = useMemo(() => {
        if (settings) {
            try {
                return buildEditorUrl(
                    `${repoName.split('/').pop() ?? ''}/${filePath}`,
                    range,
                    {
                        editorId: 'vscode',
                        projectsPaths: { default: '/Users/veszelovszki/go/src/github.com/sourcegraph' },
                    },
                    props.platformContext.sourcegraphURL
                )
            } catch {
                // TODO: Swallowing errors this way is not nice
            }
        }

        return undefined
    }, [filePath, props.platformContext.sourcegraphURL, range, repoName, settings])

    const assetsRoot = props.assetsRoot ?? (window.context?.assetsRoot || '')
    const editor = editorUrl
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
        if (editorUrl) {
            alert(`Opening ${editorUrl}`)
        } else {
            alert(`Opening setup popover. Btw, editor URL is this: ${editorUrl}`)
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
