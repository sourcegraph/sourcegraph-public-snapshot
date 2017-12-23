import * as H from 'history'
import * as React from 'react'
import { distinctUntilChanged } from 'rxjs/operators/distinctUntilChanged'
import { filter } from 'rxjs/operators/filter'
import { Subject } from 'rxjs/Subject'
import { Subscription } from 'rxjs/Subscription'
import { SaveToolbar } from '../components/SaveToolbar'
import { eventLogger } from '../tracking/eventLogger'
import { MonacoSettingsEditor } from './MonacoSettingsEditor'

interface Props {
    history: H.History

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

const emptySettings = '{\n  // add settings here (Cmd/Ctrl+Space to see hints)\n}'

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

    public componentDidMount(): void {
        // Prevent navigation when dirty.
        this.subscriptions.add(
            this.props.history.block((location: H.Location, action: H.Action) => {
                if (action === 'REPLACE') {
                    return undefined
                }
                if (this.state.saving || this.dirty) {
                    return 'Discard settings changes?'
                }
                return undefined // allow navigation
            })
        )
    }

    public componentWillReceiveProps(newProps: Props): void {
        this.componentUpdates.next(newProps)
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
            <div className="settings-file">
                <h3>Configuration</h3>
                <SaveToolbar
                    dirty={dirty}
                    disabled={this.state.saving || !dirty}
                    saving={this.state.saving}
                    onSave={this.save}
                    onDiscard={this.discard}
                />
                <MonacoSettingsEditor
                    className="settings-file__contents form-control"
                    value={contents}
                    jsonSchema="https://sourcegraph.com/v1/settings.schema.json#"
                    onChange={this.onEditorChange}
                    readOnly={this.state.saving}
                />
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
