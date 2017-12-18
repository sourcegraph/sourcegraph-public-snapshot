import CheckmarkIcon from '@sourcegraph/icons/lib/Checkmark'
import CloseIcon from '@sourcegraph/icons/lib/Close'
import ErrorIcon from '@sourcegraph/icons/lib/Error'
import Loader from '@sourcegraph/icons/lib/Loader'
import * as React from 'react'
import { distinctUntilChanged } from 'rxjs/operators/distinctUntilChanged'
import { filter } from 'rxjs/operators/filter'
import { Subject } from 'rxjs/Subject'
import { Subscription } from 'rxjs/Subscription'
import { eventLogger } from '../tracking/eventLogger'
import { MonacoSettingsEditor } from './MonacoSettingsEditor'

interface Props {
    settings: GQL.ISettings | null

    /**
     * Called when the user saves changes to the settings file's contents.
     */
    onDidCommit: (lastKnownSettingsID: number | null, contents: string) => void

    /**
     * The error that occurred on the last call to the onDidCommit callback,
     * if any.
     */
    commitError?: Error
}

interface State {
    contents?: string
    saving: boolean

    /**
     * The lastKnownSettingsID that we started editing from. If null, then no
     * previous versions of the settings exist, and we're creating them from
     * scratch.
     */
    editingLastKnownSettingsID?: number | null
}

const useMonacoEditor = () => window.localStorage.getItem('monacoSettingsEditor') !== null

const emptySettings = useMonacoEditor()
    ? '{\n  // add configuration here (Cmd/Ctrl+Space to see hints)\n}'
    : '{\n  // add configuration here\n}'

export class SettingsFile extends React.PureComponent<Props, State> {
    private componentUpdates = new Subject<Props>()
    private subscriptions = new Subscription()

    constructor(props: Props) {
        super(props)

        this.state = {
            saving: false,
            contents: this.getPropsSettingsContentsOrEmpty(),
        }

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
                        a.settings.configuration.contents === b.settings.configuration.contents &&
                        a.settings.id === b.settings.id)
            ),
            filter(
                ({ settings, commitError }) =>
                    !!settings &&
                    !commitError &&
                    ((typeof this.state.editingLastKnownSettingsID === 'number' &&
                        settings.id > this.state.editingLastKnownSettingsID) ||
                        (typeof settings.id === 'number' && this.state.editingLastKnownSettingsID === null))
            )
        )
        this.subscriptions.add(
            refreshedAfterSave.subscribe(({ settings }) =>
                this.setState({
                    saving: false,
                    editingLastKnownSettingsID: undefined,
                    contents: this.getPropsSettingsContentsOrEmpty(settings),
                })
            )
        )
    }

    public componentWillReceiveProps(newProps: Props): void {
        this.componentUpdates.next(newProps)
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        const dirty =
            this.state.contents !== undefined && this.state.contents !== this.getPropsSettingsContentsOrEmpty()
        const contents =
            this.state.contents === undefined ? this.getPropsSettingsContentsOrEmpty() : this.state.contents

        return (
            <div className="settings-file">
                <h3>Configuration</h3>
                <div className="settings-file__actions">
                    <button
                        disabled={this.state.saving || !dirty}
                        className="btn btn-sm btn-link settings-file__action"
                        onClick={this.save}
                    >
                        <CheckmarkIcon className="icon-inline" /> Save
                    </button>
                    <button
                        disabled={this.state.saving || !dirty}
                        className="btn btn-sm btn-link settings-file__action"
                        onClick={this.discard}
                    >
                        <CloseIcon className="icon-inline" /> Discard
                    </button>
                    {this.state.saving && (
                        <span className="settings-file__action">
                            <Loader className="icon-inline" /> Saving...
                        </span>
                    )}
                </div>
                {this.props.commitError && (
                    <div className="settings-file__error">
                        <ErrorIcon className="icon-inline settings-file__error-icon" />
                        {this.props.commitError.message}
                    </div>
                )}
                {useMonacoEditor() ? (
                    <MonacoSettingsEditor
                        className="settings-file__contents form-control"
                        value={contents}
                        onChange={this.onEditorChange}
                        readOnly={this.state.saving}
                    />
                ) : (
                    <code>
                        <textarea
                            className="settings-file__contents form-control"
                            value={contents}
                            onChange={this.onTextareaChange}
                            disabled={this.state.saving}
                            spellCheck={false}
                        />
                    </code>
                )}
            </div>
        )
    }

    private getPropsSettingsContentsOrEmpty(settings = this.props.settings): string {
        return settings ? settings.configuration.contents : emptySettings
    }

    private getPropsSettingsID(): number | null {
        return this.props.settings ? this.props.settings.id : null
    }

    private discard = () => {
        if (this.getPropsSettingsContentsOrEmpty() === this.state.contents || window.confirm('Really discard edits?')) {
            eventLogger.log('SettingsFileDiscard')
            this.setState({
                contents: undefined,
                editingLastKnownSettingsID: undefined,
            })
        } else {
            eventLogger.log('SettingsFileDiscardCanceled')
        }
    }

    private onTextareaChange: React.ChangeEventHandler<HTMLTextAreaElement> = event => {
        this.onEditorChange(event.target.value)
    }

    private onEditorChange = (newValue: string) => {
        if (newValue !== this.getPropsSettingsContentsOrEmpty()) {
            this.setState({ editingLastKnownSettingsID: this.getPropsSettingsID() })
        }
        this.setState({ contents: newValue })
    }

    private save = () => {
        eventLogger.log('SettingsFileSaved')
        this.setState({ saving: true }, () => {
            this.props.onDidCommit(this.getPropsSettingsID(), this.state.contents!)
        })
    }
}
