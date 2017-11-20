import CheckmarkIcon from '@sourcegraph/icons/lib/Checkmark'
import CloseIcon from '@sourcegraph/icons/lib/Close'
import ErrorIcon from '@sourcegraph/icons/lib/Error'
import Loader from '@sourcegraph/icons/lib/Loader'
import PencilIcon from '@sourcegraph/icons/lib/Pencil'
import * as React from 'react'
import { distinctUntilChanged } from 'rxjs/operators/distinctUntilChanged'
import { filter } from 'rxjs/operators/filter'
import { Subject } from 'rxjs/Subject'
import { Subscription } from 'rxjs/Subscription'
import stripJSONComments from 'strip-json-comments'
import { eventLogger } from '../tracking/eventLogger'

interface Props {
    settings: GQL.ISettings

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
    editing: boolean
    modifiedContents?: string
    saving: boolean
    editingLastKnownSettingsID?: number
    inputError?: Error
}

export class SettingsFile extends React.PureComponent<Props, State> {
    public state: State = { editing: false, saving: false }

    private componentUpdates = new Subject<Props>()
    private subscriptions = new Subscription()

    constructor(props: Props) {
        super(props)

        // We are finished saving when we receive the new settings ID and it's
        // higher than the one we saved on top of.
        const refreshedAfterSave = this.componentUpdates.pipe(
            distinctUntilChanged(
                (a, b) =>
                    a.settings.configuration.contents === b.settings.configuration.contents &&
                    a.settings.id === b.settings.id
            ),
            filter(
                ({ settings, commitError }) =>
                    !commitError &&
                    typeof this.state.editingLastKnownSettingsID === 'number' &&
                    settings.id > this.state.editingLastKnownSettingsID
            )
        )
        this.subscriptions.add(
            refreshedAfterSave.subscribe(() =>
                this.setState({
                    editing: false,
                    saving: false,
                    editingLastKnownSettingsID: undefined,
                    inputError: undefined,
                    modifiedContents: undefined,
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
        return (
            <div className="settings-file">
                <div className="settings-file__actions">
                    {this.state.saving && (
                        <span className="btn btn-icon">
                            <Loader className="icon-inline" /> Saving...
                        </span>
                    )}
                    {!this.state.saving &&
                        !this.state.editing && (
                            <button className="btn btn-icon" onClick={this.edit}>
                                <PencilIcon className="icon-inline" /> Edit
                            </button>
                        )}
                    {!this.state.saving &&
                        this.state.editing && (
                            <button className="btn btn-icon" onClick={this.save} disabled={!!this.state.inputError}>
                                <CheckmarkIcon className="icon-inline" /> Save
                            </button>
                        )}
                    {!this.state.saving &&
                        this.state.editing && (
                            <button className="btn btn-icon" onClick={this.discard}>
                                <CloseIcon className="icon-inline" /> Discard
                            </button>
                        )}
                </div>
                {this.props.commitError && (
                    <div className="settings-file__error">
                        <ErrorIcon className="icon-inline" />
                        {this.props.commitError.message}
                    </div>
                )}
                {this.state.editing &&
                    this.state.inputError && (
                        <div className="settings-file__error">
                            <ErrorIcon className="icon-inline" />
                            {this.state.inputError.message}
                        </div>
                    )}
                {this.state.editing && (
                    <textarea
                        className="settings-file__contents settings-file__contents-editor"
                        value={this.state.modifiedContents}
                        onChange={this.onTextareaChange}
                        onBlur={this.onTextareaBlur}
                        disabled={this.state.saving}
                        spellCheck={false}
                    />
                )}
                {!this.state.editing && (
                    <div
                        className="settings-file__contents"
                        dangerouslySetInnerHTML={{ __html: this.props.settings.configuration.highlighted }}
                    />
                )}
            </div>
        )
    }

    private edit = () =>
        this.setState({
            editing: true,
            modifiedContents: this.props.settings.configuration.contents,
            editingLastKnownSettingsID: this.props.settings.id,
        })

    private discard = () => {
        if (
            this.props.settings.configuration.contents === this.state.modifiedContents ||
            window.confirm('Really discard edits?')
        ) {
            this.setState({
                editing: false,
                modifiedContents: undefined,
                inputError: undefined,
            })
        }
    }

    private onTextareaChange: React.ChangeEventHandler<HTMLTextAreaElement> = event => {
        this.setState({ modifiedContents: event.target.value })

        // Clear input errors if the user's input fixed them, but (to avoid being
        // annoying) do not show new errors as the user types.
        try {
            tryParseJSONWithComments(event.target.value)
            this.setState({ inputError: undefined })
        } catch (err) {
            /* noop */
        }
    }

    private onTextareaBlur: React.FocusEventHandler<HTMLTextAreaElement> = event => {
        if (!this.state.modifiedContents) {
            return
        }

        // Format the JSON value if it has no comments. Only report an error if the parse
        // error is not due to it just having comments.
        try {
            this.setState({
                inputError: undefined,
                modifiedContents: JSON.stringify(JSON.parse(this.state.modifiedContents), null, 2),
            })
        } catch (err) {
            this.setState({
                inputError: tryParseJSONWithComments(this.state.modifiedContents),
            })
        }
    }

    private save = () => {
        eventLogger.log('SettingsFileSaved')
        this.setState({ saving: true }, () => {
            this.props.onDidCommit(this.props.settings.id, this.state.modifiedContents!)
        })
    }
}

function tryParseJSONWithComments(input: string): Error | undefined {
    try {
        JSON.parse(stripJSONComments(input))
        return undefined
    } catch (err) {
        return err
    }
}
