import * as React from 'react'

import classNames from 'classnames'
import { Subject, Subscription } from 'rxjs'
import { distinctUntilChanged, filter, map, startWith } from 'rxjs/operators'

import { ErrorLike } from '@sourcegraph/common'
import { Settings } from '@sourcegraph/shared/src/schema/settings.schema'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { BeforeUnloadPrompt, LoadingSpinner, Tab, TabList, TabPanel, TabPanels, Tabs } from '@sourcegraph/wildcard'

import settingsSchemaJSON from '../../../../schema/settings.schema.json'
import { SaveToolbar } from '../components/SaveToolbar'
import { SiteAdminSettingsCascadeFields } from '../graphql-operations'
import { eventLogger } from '../tracking/eventLogger'

import { GeneratedSettingsForm, SettingsNode } from './GeneratedSettingsForm'

import styles from './SettingsFile.module.scss'

interface Props extends TelemetryProps {
    settings: SiteAdminSettingsCascadeFields['subjects'][number]['latestSettings'] | null

    settingsCascadeFinal: Settings | ErrorLike | null

    /**
     * Called when the user saves changes to the settings file's contents.
     */
    onDidCommit: (lastID: number | null, contents: string) => void

    /**
     * Called when the user discards changes to the settings file's contents.
     */
    onDidDiscard: () => void

    /**
     * The error that occurred on the last call to the onDidCommit callback,
     * if any.
     */
    commitError?: Error

    isLightTheme: boolean
}

interface State {
    contents?: string
    saving: boolean
    isFormDirty: boolean

    /**
     * The lastID that we started editing from. If null, then no
     * previous versions of the settings exist, and we're creating them from
     * scratch.
     */
    editingLastID?: number | null
}

const emptySettings = '{\n  // add settings here (Ctrl+Space to see hints)\n}'

const MonacoSettingsEditor = React.lazy(async () => ({
    default: (await import('./MonacoSettingsEditor')).MonacoSettingsEditor,
}))

export class SettingsFile extends React.PureComponent<Props, State> {
    private componentUpdates = new Subject<Props>()
    private subscriptions = new Subscription()

    constructor(props: Props) {
        super(props)

        this.state = { saving: false, isFormDirty: false }

        // Reset state upon navigation to a different subject.
        this.subscriptions.add(
            this.componentUpdates
                .pipe(
                    startWith(props),
                    map(({ settings }) => settings),
                    distinctUntilChanged()
                )
                .subscribe(() => {
                    if (this.state.contents !== undefined) {
                        this.setState({ contents: undefined })
                    }
                })
        )

        // Saving ended (in failure) if we get a commitError.
        this.subscriptions.add(
            this.componentUpdates
                .pipe(
                    map(({ commitError }) => commitError),
                    distinctUntilChanged(),
                    filter(commitError => !!commitError)
                )
                .subscribe(() => this.setState({ saving: false }))
        )

        // We are finished saving when we receive the new settings ID and it's
        // higher than the one we saved on top of.
        const refreshedAfterSave = this.componentUpdates.pipe(
            filter(({ settings }) => !!settings),
            distinctUntilChanged(
                (a, b) =>
                    (!a.settings && !!b.settings) ||
                    (!!a.settings && !b.settings) ||
                    (!!a.settings &&
                        !!b.settings &&
                        a.settings.contents === b.settings.contents &&
                        a.settings.id === b.settings.id)
            ),
            filter(
                ({ settings, commitError }) =>
                    !!settings &&
                    !commitError &&
                    ((typeof this.state.editingLastID === 'number' && settings.id > this.state.editingLastID) ||
                        (typeof settings.id === 'number' && this.state.editingLastID === null))
            )
        )
        this.subscriptions.add(
            refreshedAfterSave.subscribe(({ settings }) =>
                this.setState({
                    saving: false,
                    editingLastID: undefined,
                    contents: settings ? settings.contents : undefined,
                })
            )
        )
    }

    public componentDidUpdate(): void {
        this.componentUpdates.next(this.props)
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    private get isJSONEditorDirty(): boolean {
        return this.state.contents !== undefined && this.state.contents !== this.getPropsSettingsContentsOrEmpty()
    }

    public render(): JSX.Element | null {
        const contents =
            this.state.contents === undefined ? this.getPropsSettingsContentsOrEmpty() : this.state.contents

        return (
            <div className={classNames('test-settings-file d-flex flex-grow-1 flex-column', styles.settingsFile)}>
                <Tabs>
                    <TabList>
                        <Tab disabled={false /* TODO: Make this dynamic */}>Settings</Tab>
                        <Tab disabled={this.state.isFormDirty}>JSON Editor</Tab>
                    </TabList>
                    <TabPanels>
                        <TabPanel>
                            <GeneratedSettingsForm
                                jsonSchema={settingsSchemaJSON as unknown as SettingsNode}
                                currentSettings={this.props.settingsCascadeFinal}
                                reportDirtiness={this.onFormChange}
                            />
                        </TabPanel>
                        <TabPanel>
                            <BeforeUnloadPrompt when={this.state.saving || this.dirty} message="Discard settings changes?" />
                            <React.Suspense fallback={<LoadingSpinner className="mt-2" />}>
                                <MonacoSettingsEditor
                                    value={contents}
                                    jsonSchema={settingsSchemaJSON}
                                    onChange={this.onJSONEditorChange}
                                    readOnly={this.state.saving}
                                    isLightTheme={this.props.isLightTheme}
                                    onDidSave={this.saveJSON}
                                />
                            </React.Suspense>
                            <SaveToolbar
                                dirty={this.isJSONEditorDirty}
                                error={this.props.commitError}
                                saving={this.state.saving}
                                onSave={this.saveJSON}
                                onDiscard={this.discardJSON}
                            />
                        </TabPanel>
                    </TabPanels>
                </Tabs>
            </div>
        )
    }

    private getPropsSettingsContentsOrEmpty(settings = this.props.settings): string {
        return settings ? settings.contents : emptySettings
    }

    private getPropsSettingsID(): number | null {
        return this.props.settings ? this.props.settings.id : null
    }

    private onFormChange = (isDirty: boolean): void => {
        this.setState({ isFormDirty: isDirty })
    }

    private discardJSON = (): void => {
        if (
            this.getPropsSettingsContentsOrEmpty() === this.state.contents ||
            window.confirm('Discard settings edits?')
        ) {
            eventLogger.log('SettingsFileDiscard')
            this.setState({
                contents: undefined,
                editingLastID: undefined,
            })
            this.props.onDidDiscard()
        } else {
            eventLogger.log('SettingsFileDiscardCanceled')
        }
    }

    private onJSONEditorChange = (newValue: string): void => {
        if (newValue !== this.getPropsSettingsContentsOrEmpty()) {
            this.setState({ editingLastID: this.getPropsSettingsID() })
        }
        this.setState({ contents: newValue })
    }

    private saveJSON = (): void => {
        eventLogger.log('SettingsFileSaved')
        this.setState({ saving: true }, () => {
            this.props.onDidCommit(this.getPropsSettingsID(), this.state.contents!)
        })
    }
}
