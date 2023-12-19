import * as React from 'react'
import { useCallback, useEffect, useMemo, useState } from 'react'

import { mdiApplicationEditOutline } from '@mdi/js'
import { from } from 'rxjs'

import { logger } from '@sourcegraph/common'
import { SimpleActionItem } from '@sourcegraph/shared/src/actions/SimpleActionItem'
import type { PlatformContext } from '@sourcegraph/shared/src/platform/context'
import { isSettingsValid, type Settings } from '@sourcegraph/shared/src/settings/settings'
import {
    Button,
    Icon,
    Popover,
    PopoverContent,
    PopoverTrigger,
    Position,
    Tooltip,
    useObservable,
} from '@sourcegraph/wildcard'

import { RepoHeaderActionAnchor, RepoHeaderActionMenuLink } from '../repo/components/RepoHeaderActions'
import { RepoActionInfo } from '../repo/RepoActionInfo'
import { eventLogger } from '../tracking/eventLogger'

import { getEditorSettingsErrorMessage } from './build-url'
import type { EditorSettings } from './editor-settings'
import { type EditorId, getEditor } from './editors'
import { migrateLegacySettings } from './migrate-legacy-settings'
import { OpenInEditorPopover } from './OpenInEditorPopover'
import { useOpenCurrentUrlInEditor } from './useOpenCurrentUrlInEditor'

import styles from './OpenInEditorActionItem.module.scss'

export interface OpenInEditorActionItemProps {
    platformContext: PlatformContext
    externalServiceType?: string
    assetsRoot?: string
    source?: 'repoHeader' | 'actionItemsBar'
    actionType?: 'nav' | 'dropdown'
}

// We only want to attempt to upgrade the legacy open in editor settings once per
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

    const shouldShowEditorText = editors?.length === 1
    return editors ? (
        <>
            {editors.map(
                (editor, index) =>
                    editor && (
                        <EditorItem
                            key={editor.id}
                            tooltip={`Open file in ${editor?.name}`}
                            icon={
                                <img
                                    src={`${assetsRoot}/img/editors/${editor.id}.svg`}
                                    alt={`Open file in ${editor?.name}`}
                                    className={styles.icon}
                                    // className={classNames(styles.icon, styles.repoActionIcon)}
                                />
                            }
                            onClick={() => {
                                eventLogger.log('OpenInEditorClicked', { editor: editor.id }, { editor: editor.id })
                                openCurrentUrlInEditor(
                                    settings?.openInEditor,
                                    props.externalServiceType,
                                    props.platformContext.sourcegraphURL,
                                    index
                                )
                            }}
                            source={props.source}
                            actionType={props.actionType}
                            shouldShowEditorText={shouldShowEditorText}
                        />
                    )
            )}
        </>
    ) : // We can not render the editor popover inside the dropdown view yet.
    // Since the dropdown view is only used on very limited viewport dimensions,
    // we're okay with having the user go to the settings to configure this
    // instead.
    //
    // Chances are that they won't have an IDE configured on these devices
    // anyways.
    props.actionType !== 'dropdown' ? (
        <Popover isOpen={popoverOpen} onOpenChange={event => setPopoverOpen(event.isOpen)}>
            <PopoverTrigger as="div" className={styles.item}>
                <EditorItem
                    tooltip="Set your preferred editor"
                    isActive={popoverOpen}
                    icon={
                        props.source === 'repoHeader' ? (
                            <Icon aria-hidden={true} className={styles.icon} svgPath={mdiApplicationEditOutline} />
                        ) : (
                            <img
                                src={`${assetsRoot}/img/open-in-editor.svg`}
                                alt="Set your preferred editor"
                                className={styles.icon}
                            />
                        )
                    }
                    onClick={togglePopover}
                    source={props.source}
                    actionType={props.actionType}
                    shouldShowEditorText={true}
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
    ) : null
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
                logger.log('Migrated items successfully.')
            })
            .catch(error => {
                logger.error('Setting migration failed.', error)
            })
    }
}

interface EditorItemProps {
    tooltip: string
    onClick: () => void
    icon: React.ReactNode
    isActive?: boolean
    source?: 'repoHeader' | 'actionItemsBar'
    actionType?: 'nav' | 'dropdown'
    shouldShowEditorText?: boolean
}

function EditorItem(props: EditorItemProps): JSX.Element {
    if (props.source === 'actionItemsBar') {
        return (
            <SimpleActionItem tooltip={props.tooltip} onSelect={props.onClick} isActive={props.isActive}>
                {props.icon}
            </SimpleActionItem>
        )
    }

    if (props.actionType === 'dropdown') {
        return (
            <RepoHeaderActionMenuLink file={true} as={Button} onClick={props.onClick}>
                {props.icon}
                <span>{props.tooltip}</span>
            </RepoHeaderActionMenuLink>
        )
    }

    return (
        <Tooltip content={props.tooltip}>
            <RepoHeaderActionAnchor onSelect={props.onClick} className={styles.item}>
                <RepoActionInfo icon={props.icon} displayName="Editor" hideActionLabel={!props.shouldShowEditorText} />
            </RepoHeaderActionAnchor>
        </Tooltip>
    )
}
