import * as React from 'react'
import { useCallback, useEffect, useState } from 'react'

import { Unsubscribable } from 'rxjs'

import { isErrorLike } from '@sourcegraph/common'
import { PlatformContext } from '@sourcegraph/shared/out/src/platform/context'
import { SettingsCascadeOrError } from '@sourcegraph/shared/out/src/settings/settings'
import { Popover, PopoverContent, PopoverTrigger, Position } from '@sourcegraph/wildcard'

import { SimpleActionItem } from '../../../shared/src/actions/SimpleActionItem'

import { getEditorSettingsErrorMessage } from './build-url'
import type { EditorSettings } from './editor-settings'
import { EditorId, getEditor } from './editors'
import { migrateLegacySettings } from './migrate-legacy-settings'
import { OpenInEditorPopover } from './OpenInEditorPopover'
import { useOpenCurrentUrlInEditor } from './useOpenCurrentUrlInEditor'

export interface OpenInEditorActionItemProps {
    platformContext: PlatformContext
    assetsRoot?: string
}

export const OpenInEditorActionItem: React.FunctionComponent<OpenInEditorActionItemProps> = props => {
    const assetsRoot = props.assetsRoot ?? (window.context?.assetsRoot || '')

    const [settingsCascadeOrError, setSettingsCascadeOrError] = useState<SettingsCascadeOrError | undefined>(undefined)
    const settings = !isErrorLike(settingsCascadeOrError?.final) ? settingsCascadeOrError?.final : undefined
    const [settingSubscription, setSettingSubscription] = useState<Unsubscribable | null>(null)
    const userSettings = settingsCascadeOrError?.subjects
        ? settingsCascadeOrError.subjects[settingsCascadeOrError.subjects.length - 1]
        : undefined

    const [popoverOpen, setPopoverOpen] = useState(false)
    const togglePopover = useCallback(() => {
        setPopoverOpen(previous => !previous)
    }, [])

    const openCurrentUrlInEditor = useOpenCurrentUrlInEditor()

    const editorSettingsErrorMessage = getEditorSettingsErrorMessage(
        settings?.openInEditor,
        props.platformContext.sourcegraphURL
    )
    const editor = !editorSettingsErrorMessage
        ? getEditor((settings?.openInEditor as EditorSettings | undefined)?.editorIds?.[0] || '')
        : undefined

    useEffect(() => {
        setSettingSubscription(
            props.platformContext.settings.subscribe(settings => {
                if (settings.final) {
                    /* Migrate legacy settings if needed */
                    const subject = settings.subjects ? settings.subjects[settings.subjects.length - 1] : undefined
                    if (subject?.settings && !isErrorLike(subject.settings) && !subject.settings.openInEditor) {
                        const migratedSettings = migrateLegacySettings(subject.settings)
                        props.platformContext
                            .updateSettings(subject.subject.id, JSON.stringify(migratedSettings, null, 4))
                            .then(() => {
                                console.log('Migrated items successfully.')
                            })
                            .catch(() => {
                                // TODO: Update failed, handle this later
                            })
                    }
                    setSettingsCascadeOrError(settings)
                }
            })
        )

        return () => {
            settingSubscription?.unsubscribe()
        }
    }, [settingSubscription, props.platformContext.settings, props.platformContext])

    const onClick = useCallback(
        (event: React.MouseEvent<HTMLElement>) => {
            event.stopPropagation()
            if (editor) {
                openCurrentUrlInEditor(settings?.openInEditor, props.platformContext.sourcegraphURL)
            } else {
                togglePopover()
            }
        },
        [editor, openCurrentUrlInEditor, props.platformContext.sourcegraphURL, settings?.openInEditor, togglePopover]
    )

    const onSave = useCallback(
        async (selectedEditorId: EditorId, defaultProjectPath: string): Promise<void> => {
            if (!userSettings) {
                throw new Error('No user settings. Not saving.')
            }
            await props.platformContext.updateSettings(userSettings.subject.id, {
                path: ['openInEditor', 'projectPaths.default'],
                value: defaultProjectPath,
            })
            await props.platformContext.updateSettings(userSettings.subject.id, {
                path: ['openInEditor', 'editorIds'],
                value: [selectedEditorId],
            })
        },
        [props.platformContext, userSettings]
    )

    return (
        <Popover isOpen={popoverOpen} onOpenChange={event => setPopoverOpen(event.isOpen)}>
            <PopoverTrigger as="div">
                <SimpleActionItem
                    tooltip={editor ? `Open file in ${editor?.name}` : 'Set your preferred editor'}
                    isActive={popoverOpen}
                    iconURL={
                        editor ? `${assetsRoot}/img/editors/${editor.id}.svg` : `${assetsRoot}/img/open-in-editor.svg`
                    }
                    onClick={onClick}
                />
            </PopoverTrigger>
            <PopoverContent position={Position.leftStart} className="pt-0 pb-0" aria-labelledby="repo-revision-popover">
                <OpenInEditorPopover
                    editorSettings={settings?.openInEditor as EditorSettings | undefined}
                    togglePopover={togglePopover}
                    onSave={onSave}
                    sourcegraphUrl={props.platformContext.sourcegraphURL}
                />
            </PopoverContent>
        </Popover>
    )
}
