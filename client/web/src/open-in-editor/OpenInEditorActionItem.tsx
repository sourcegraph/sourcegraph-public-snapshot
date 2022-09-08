import * as React from 'react'
import { useCallback, useEffect, useState } from 'react'

import { Unsubscribable } from 'rxjs'

import { isErrorLike } from '@sourcegraph/common'
import { SimpleActionItem } from '@sourcegraph/shared/src/actions/SimpleActionItem'
import { PlatformContext } from '@sourcegraph/shared/src/platform/context'
import { SettingsCascadeOrError } from '@sourcegraph/shared/src/settings/settings'
import { Popover, PopoverContent, PopoverTrigger, Position } from '@sourcegraph/wildcard'

import { eventLogger } from '../tracking/eventLogger'

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
    const editorIds = (settings?.openInEditor as EditorSettings | undefined)?.editorIds ?? []
    const editors = !editorSettingsErrorMessage ? editorIds.map(getEditor) : undefined

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

    return editors ? (
        <>
            {editors.map(
                (editor, index) =>
                    editor && (
                        <SimpleActionItem
                            key={editor.id}
                            tooltip={`Open file in ${editor?.name}`}
                            iconURL={`${assetsRoot}/img/editors/${editor.id}.svg`}
                            onClick={() => {
                                eventLogger.log('OpenInEditorClicked', { editor: editor.id }, { editor: editor.id })
                                openCurrentUrlInEditor(
                                    settings?.openInEditor,
                                    props.platformContext.sourcegraphURL,
                                    index
                                )
                            }}
                        />
                    )
            )}
        </>
    ) : (
        <Popover isOpen={popoverOpen} onOpenChange={event => setPopoverOpen(event.isOpen)}>
            <PopoverTrigger as="div">
                <SimpleActionItem
                    tooltip="Set your preferred editor"
                    isActive={popoverOpen}
                    iconURL={`${assetsRoot}/img/open-in-editor.svg`}
                    onClick={togglePopover}
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
