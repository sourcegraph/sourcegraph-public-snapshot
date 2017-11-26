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
import { eventLogger } from '../tracking/eventLogger'

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
    editing: boolean
    modifiedContents?: string
    saving: boolean

    /**
     * The lastKnownSettingsID that we started editing from. If null, then no
     * previous versions of the settings exist, and we're creating them from
     * scratch.
     */
    editingLastKnownSettingsID?: number | null
}

const emptySettings = '{\n  // empty configuration\n}'

export class SettingsFile extends React.PureComponent<Props, State> {
    public state: State = { editing: false, saving: false }

    private componentUpdates = new Subject<Props>()
    private subscriptions = new Subscription()

    constructor(props: Props) {
        super(props)

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
            refreshedAfterSave.subscribe(() =>
                this.setState({
                    editing: false,
                    saving: false,
                    editingLastKnownSettingsID: undefined,
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
                        <span className="btn btn-icon settings-file__actions-btn">
                            <Loader className="icon-inline" /> Saving...
                        </span>
                    )}
                    {!this.state.saving &&
                        !this.state.editing && (
                            <button className="btn btn-icon settings-file__actions-btn" onClick={this.edit}>
                                <PencilIcon className="icon-inline" /> Edit
                            </button>
                        )}
                    {!this.state.saving &&
                        this.state.editing && (
                            <button className="btn btn-icon settings-file__actions-btn" onClick={this.save}>
                                <CheckmarkIcon className="icon-inline" /> Save
                            </button>
                        )}
                    {!this.state.saving &&
                        this.state.editing && (
                            <button className="btn btn-icon settings-file__actions-btn" onClick={this.discard}>
                                <CloseIcon className="icon-inline" /> Discard
                            </button>
                        )}
                </div>
                {this.props.commitError && (
                    <div className="settings-file__error">
                        <ErrorIcon className="icon-inline settings-file__error-icon" />
                        {this.props.commitError.message}
                    </div>
                )}
                {this.state.editing && (
                    <textarea
                        className="settings-file__contents settings-file__contents-editor"
                        value={this.state.modifiedContents}
                        onChange={this.onTextareaChange}
                        disabled={this.state.saving}
                        spellCheck={false}
                    />
                )}
                {!this.state.editing &&
                    (this.props.settings ? (
                        <div
                            className="settings-file__contents"
                            dangerouslySetInnerHTML={{ __html: this.props.settings.configuration.highlighted }}
                        />
                    ) : (
                        <div className="settings-file__contents">{emptySettings}</div>
                    ))}
            </div>
        )
    }

    private getPropsSettingsContentsOrEmpty(): string {
        return this.props.settings ? this.props.settings.configuration.contents : emptySettings
    }

    private getPropsSettingsID(): number | null {
        return this.props.settings ? this.props.settings.id : null
    }

    private edit = () =>
        this.setState({
            editing: true,
            modifiedContents: this.getPropsSettingsContentsOrEmpty(),
            editingLastKnownSettingsID: this.getPropsSettingsID(),
        })

    private discard = () => {
        if (
            this.getPropsSettingsContentsOrEmpty() === this.state.modifiedContents ||
            window.confirm('Really discard edits?')
        ) {
            this.setState({
                editing: false,
                modifiedContents: undefined,
            })
        }
    }

    private onTextareaChange: React.ChangeEventHandler<HTMLTextAreaElement> = event => {
        this.setState({ modifiedContents: event.target.value })
    }

    private save = () => {
        eventLogger.log('SettingsFileSaved')
        this.setState({ saving: true }, () => {
            this.props.onDidCommit(this.getPropsSettingsID(), this.state.modifiedContents!)
        })
    }
}
