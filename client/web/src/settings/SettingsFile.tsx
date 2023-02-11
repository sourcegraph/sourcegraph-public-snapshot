import * as React from 'react'

import classNames from 'classnames'
// eslint-disable-next-line no-restricted-syntax
import { useHistory } from 'react-router'

import { ErrorLike } from '@sourcegraph/common'
import { Settings } from '@sourcegraph/shared/src/schema/settings.schema'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { BeforeUnloadPrompt, LoadingSpinner, Tab, TabList, TabPanel, TabPanels, Tabs } from '@sourcegraph/wildcard'

import settingsSchemaJSON from '../../../../schema/settings.schema.json'
import { SaveToolbar } from '../components/SaveToolbar'
import { useFeatureFlag } from '../featureFlags/useFeatureFlag'
import { SiteAdminSettingsCascadeFields } from '../graphql-operations'
import { eventLogger } from '../tracking/eventLogger'

import { GeneratedSettingsForm, SettingsNode } from './GeneratedSettingsForm'

import styles from './SettingsFile.module.scss'

interface Props extends TelemetryProps {
    settings: SiteAdminSettingsCascadeFields['subjects'][number]['latestSettings'] | null
    settingsCascadeFinal: Settings | ErrorLike | null
    // Called when the user saves changes to the settings file's contents.
    onDidCommit: (lastID: number | null, contents: string) => void
    // Called when the user discards changes to the settings file's contents.
    onDidDiscard: () => void
    // The error that occurred on the last call to the onDidCommit callback, if any.
    commitError?: Error

    isLightTheme: boolean
}

const emptySettings = '{\n  // add settings here (Ctrl+Space to see hints)\n}'

const MonacoSettingsEditor = React.lazy(async () => ({
    default: (await import('./MonacoSettingsEditor')).MonacoSettingsEditor,
}))

export const SettingsFile = ({
    settings,
    settingsCascadeFinal,
    onDidCommit,
    onDidDiscard,
    commitError,
    isLightTheme,
}: Props): JSX.Element => {
    const [contents, setContents] = React.useState<string | undefined>(settings ? settings.contents : emptySettings)
    const [saving, setSaving] = React.useState(false)
    const [isFormDirty, setIsFormDirty] = React.useState(false)
    // The lastID that we started editing from. If null, then no previous versions of the settings exist, and we're creating them from scratch.
    const [editingLastID, setEditingLastID] = React.useState<number | null | undefined>(undefined)

    // eslint-disable-next-line no-restricted-syntax
    const history = useHistory()

    const enableVisualSettingsEditor = useFeatureFlag('visual-settings-editor')

    // Reset state upon navigation to a different subject.
    React.useEffect(() => {
        if (contents !== undefined) {
            setContents(undefined)
        }
        // We only want to reset the state when the settings props change, not when the current settings content changes.
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [settings])

    // Saving ended (in failure) if we get a commitError.
    React.useEffect(() => {
        if (commitError) {
            setSaving(false)
        }
    }, [commitError])

    // We are finished saving when we receive the new settings ID, and it's higher than the one we saved on top of.
    React.useEffect(() => {
        if (settings && (!editingLastID || settings.id > editingLastID) && !commitError) {
            setSaving(false)
            setEditingLastID(undefined)
            setContents(settings.contents)
        }
    }, [settings, commitError, editingLastID])

    // Prevent navigation when dirty.
    React.useEffect(() => {
        const blocker = history.block((location, action) => {
            if (action === 'REPLACE') {
                return undefined
            }
            if (saving || isFormDirty) {
                return 'Discard settings changes?'
            }
            return undefined // allow navigation
        })
        return () => blocker()
    }, [saving, isFormDirty, history])

    const jsonEditor = (
        <>
            <BeforeUnloadPrompt when={saving || isFormDirty || isJSONEditorDirty()} message="Discard settings changes?" />
            <React.Suspense fallback={<LoadingSpinner className="mt-2" />}>
                <MonacoSettingsEditor
                    value={contents}
                    jsonSchema={settingsSchemaJSON}
                    onChange={onJSONEditorChange}
                    readOnly={saving}
                    isLightTheme={isLightTheme}
                    onDidSave={saveJSON}
                />
            </React.Suspense>
            <SaveToolbar
                dirty={isJSONEditorDirty()}
                error={commitError}
                saving={saving}
                onSave={saveJSON}
                onDiscard={discardJSON}
            />
        </>
    )

    if (enableVisualSettingsEditor) {
        return (
            <div className={classNames('test-settings-file d-flex flex-grow-1 flex-column', styles.settingsFile)}>
                <Tabs>
                    <TabList>
                        <Tab disabled={false /* TODO: Make this dynamic */}>Settings</Tab>
                        <Tab disabled={isFormDirty}>JSON Editor</Tab>
                    </TabList>
                    <TabPanels>
                        <TabPanel>
                            <GeneratedSettingsForm
                                jsonSchema={settingsSchemaJSON as unknown as SettingsNode}
                                currentSettings={settingsCascadeFinal}
                                reportDirtiness={onFormChange}
                            />
                        </TabPanel>
                        <TabPanel>{jsonEditor}</TabPanel>
                    </TabPanels>
                </Tabs>
            </div>
        )
    }
    return (
        <div className={classNames('test-settings-file d-flex flex-grow-1 flex-column', styles.settingsFile)}>
            {jsonEditor}
        </div>
    )

    function isJSONEditorDirty(): boolean {
        return contents !== undefined && contents !== getPropsSettingsContentsOrEmpty()
    }

    function getPropsSettingsContentsOrEmpty(): string {
        return settings ? settings.contents : emptySettings
    }

    function getPropsSettingsID(): number | null {
        return settings ? settings.id : null
    }

    function onFormChange(isDirty: boolean): void {
        setIsFormDirty(isDirty)
    }

    function discardJSON(): void {
        if (getPropsSettingsContentsOrEmpty() === contents || window.confirm('Discard settings edits?')) {
            eventLogger.log('SettingsFileDiscard')
            setContents(undefined)
            setEditingLastID(undefined)
            onDidDiscard()
        } else {
            eventLogger.log('SettingsFileDiscardCanceled')
        }
    }

    function onJSONEditorChange(newValue: string): void {
        if (newValue !== getPropsSettingsContentsOrEmpty()) {
            setEditingLastID(getPropsSettingsID())
        }
        setContents(newValue)
    }

    function saveJSON(): void {
        eventLogger.log('SettingsFileSaved')
        setSaving(true)
        if (contents) {
            onDidCommit(getPropsSettingsID(), contents)
        }
    }
}
