import * as React from 'react'
import { useCallback, useEffect, useMemo, useState } from 'react'

import { from } from 'rxjs'

import { SimpleActionItem } from '@sourcegraph/shared/src/actions/SimpleActionItem'
import { PlatformContext } from '@sourcegraph/shared/src/platform/context'
import { isSettingsValid, Settings } from '@sourcegraph/shared/src/settings/settings'
import { Popover, PopoverContent, PopoverTrigger, Position, useObservable } from '@sourcegraph/wildcard'

import { eventLogger } from '../tracking/eventLogger'

import { getEditorSettingsErrorMessage } from './build-url'
import type { EditorSettings } from './editor-settings'
import { EditorId, getEditor } from './editors'
import { migrateLegacySettings } from './migrate-legacy-settings'
import { OpenInEditorPopover } from './OpenInEditorPopover'
import { useOpenCurrentUrlInEditor } from './useOpenCurrentUrlInEditor'

import styles from './OpenInEditorActionItem.module.scss'

export interface OpenInEditorActionItemProps {
    platformContext: PlatformContext
    assetsRoot?: string
}

// We only want to attemt to upgrade the legacy open in editor settings once per
// page load.
let didAttemptToUpgradeSettings = false

export const OpenInEditorActionItem: React.FunctionComponent<OpenInEditorActionItemProps> = props => {
    const assetsRoot = props.assetsRoot ?? (window.context?.assetsRoot || '')

    const settingsOrError = useObservable(useMemo(() => from(props.platformContext.settings), [props.platformContext]))
    const settings =
        settingsOrError !== undefined && isSettingsValid(settingsOrError) ? settingsOrError.final : undefined
    const userSettingsSubject = settingsOrError?.subjects
        ? settingsOrError?.subjects.find(subject => subject.subject.__typename === 'User')?.subject.id
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
        if (!settings || !userSettingsSubject || didAttemptToUpgradeSettings) {
            return
        }
        didAttemptToUpgradeSettings = true
        upgradeSettings(props.platformContext, settings, userSettingsSubject)
    }, [props.platformContext, settings, userSettingsSubject])

    const onSave = useCallback(
        async (selectedEditorId: EditorId, defaultProjectPath: string): Promise<void> => {
            if (!userSettingsSubject) {
                throw new Error('No user settings. Not saving.')
            }
            await props.platformContext.updateSettings(userSettingsSubject, {
                path: ['openInEditor', 'projectPaths.default'],
                value: defaultProjectPath,
            })
            await props.platformContext.updateSettings(userSettingsSubject, {
                path: ['openInEditor', 'editorIds'],
                value: [selectedEditorId],
            })
        },
        [props.platformContext, userSettingsSubject]
    )

    return editors ? (
        <>
            {editors.map(
                (editor, index) =>
                    editor && (
                        <SimpleActionItem
                            key={editor.id}
                            tooltip={`Open file in ${editor?.name}`}
                            onSelect={() => {
                                eventLogger.log('OpenInEditorClicked', { editor: editor.id }, { editor: editor.id })
                                openCurrentUrlInEditor(
                                    settings?.openInEditor,
                                    props.platformContext.sourcegraphURL,
                                    index
                                )
                            }}
                        >
                            <img
                                src={`${assetsRoot}/img/editors/${editor.id}.svg`}
                                alt={`Open file in ${editor?.name}`}
                                className={styles.icon}
                            />
                        </SimpleActionItem>
                    )
            )}
        </>
    ) : (
        <Popover isOpen={popoverOpen} onOpenChange={event => setPopoverOpen(event.isOpen)}>
            <PopoverTrigger as="div">
                <SimpleActionItem tooltip="Set your preferred editor" isActive={popoverOpen} onSelect={togglePopover}>
                    <img
                        src={`${assetsRoot}/img/open-in-editor.svg`}
                        alt="Set your preferred editor"
                        className={styles.icon}
                    />
                </SimpleActionItem>
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

function upgradeSettings(platformContext: PlatformContext, settings: Settings, userSettingsSubject: string): void {
    const openInEditor = migrateLegacySettings(settings)

    if (openInEditor !== null) {
        platformContext
            .updateSettings(userSettingsSubject, {
                path: ['openInEditor'],
                value: openInEditor,
            })
            .then(() => {
                console.log('Migrated items successfully.')
            })
            .catch(error => {
                console.error('Setting migration failed.', error)
            })
    }
}
