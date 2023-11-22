import * as React from 'react'

import classNames from 'classnames'
import { Subject, Subscription } from 'rxjs'
import { distinctUntilChanged, filter, map, startWith } from 'rxjs/operators'

import type { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { LoadingSpinner, BeforeUnloadPrompt } from '@sourcegraph/wildcard'

import settingsSchemaJSON from '../../../../schema/settings.schema.json'
import { SaveToolbar } from '../components/SaveToolbar'
import type { SiteAdminSettingsCascadeFields } from '../graphql-operations'
import { eventLogger } from '../tracking/eventLogger'

import styles from './SettingsFile.module.scss'

interface Props extends TelemetryProps {
    settings: SiteAdminSettingsCascadeFields['subjects'][number]['latestSettings'] | null

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

        this.state = { saving: false }

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

    private get dirty(): boolean {
        return this.state.contents !== undefined && this.state.contents !== this.getPropsSettingsContentsOrEmpty()
    }

    public render(): JSX.Element | null {
        const dirty = this.dirty
        const contents =
            this.state.contents === undefined ? this.getPropsSettingsContentsOrEmpty() : this.state.contents

        return (
            <div
                className={classNames(
                    'test-settings-file percy-hide d-flex flex-grow-1 flex-column',
                    styles.settingsFile
                )}
            >
                <BeforeUnloadPrompt when={this.state.saving || this.dirty} message="Discard settings changes?" />
                <React.Suspense fallback={<LoadingSpinner className="mt-2" />}>
                    <MonacoSettingsEditor
                        value={contents}
                        jsonSchema={settingsSchemaJSON}
                        onChange={this.onEditorChange}
                        readOnly={this.state.saving}
                        isLightTheme={this.props.isLightTheme}
                        onDidSave={this.save}
                    />
                </React.Suspense>
                <SaveToolbar
                    dirty={dirty}
                    error={this.props.commitError}
                    saving={this.state.saving}
                    onSave={this.save}
                    onDiscard={this.discard}
                />
            </div>
        )
    }

    private getPropsSettingsContentsOrEmpty(settings = this.props.settings): string {
        return settings ? settings.contents : emptySettings
    }

    private getPropsSettingsID(): number | null {
        return this.props.settings ? this.props.settings.id : null
    }

    private discard = (): void => {
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

    private onEditorChange = (newValue: string): void => {
        if (newValue !== this.getPropsSettingsContentsOrEmpty()) {
            this.setState({ editingLastID: this.getPropsSettingsID() })
        }
        this.setState({ contents: newValue })
    }

    private save = (): void => {
        eventLogger.log('SettingsFileSaved')
        this.setState({ saving: true }, () => {
            this.props.onDidCommit(this.getPropsSettingsID(), this.state.contents!)
        })
    }
}
