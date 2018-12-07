import * as React from 'react'
import { from, interval, Observable, of, Subject, Subscription, timer } from 'rxjs'
import { ajax } from 'rxjs/ajax'
import { delay, distinctUntilChanged, mapTo, startWith, takeUntil } from 'rxjs/operators'
import './CriticalConfigEditor.scss'
import { MonacoEditor } from './MonacoEditor'

const DEBUG_LOADING_STATE_DELAY = 0 // ms

/**
 * Amount of time to wait before showing the loading indicator.
 */
const WAIT_BEFORE_SHOWING_LOADER = 250 // ms

// TODO(slimsag): future: Show errors that occur during loading
// TODO(slimsag): future: Show errors that occur during loading
// TODO(slimsag): future: Show errors that occur during saving
// TODO(slimsag): future: Warn user if they are discarding changes
// TODO(slimsag): future: Explicit discard changes button?
// TODO(slimsag): future: Better button styling
// TODO(slimsag): future: Better link styling
// TODO(slimsag): future: Better 'loading' state styling

interface Props {}

interface State {
    /** The current config content according to the server. */
    criticalConfig: { ID: string; Contents: string } | null

    /** The current content in the editor. */
    content: string | null

    /** Whether or not the loader can be shown yet, iff criticalConfig === null */
    canShowLoader: boolean
}

export class CriticalConfigEditor extends React.Component<Props, State> {
    public state: State = {
        criticalConfig: null,
        content: null,
        canShowLoader: false,
    }

    private componentUpdates = new Subject<Props>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        const componentUpdates = this.componentUpdates.pipe(startWith(this.props))

        // Periodically rerender our component in case our request takes longer
        // than `WAIT_BEFORE_SHOWING_LOADER` and we need to show the loading
        // indicator.
        this.subscriptions.add(timer(WAIT_BEFORE_SHOWING_LOADER).subscribe(t => this.setState({ canShowLoader: true })))

        // Load the initial critical config.
        this.subscriptions.add(
            ajax('/get')
                .pipe(delay(DEBUG_LOADING_STATE_DELAY))
                .subscribe(resp => {
                    this.setState({
                        criticalConfig: resp.response,
                        content: resp.response.Contents,
                    })
                })
        )
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        return (
            <div className="critical-config-editor">
                <div
                    className={`critical-config-editor__monaco-reserved-space${
                        this.state.criticalConfig ? ' critical-config-editor__monaco-reserved-space--monaco' : ''
                    }`}
                >
                    {!this.state.criticalConfig && this.state.canShowLoader && <div>Loading...</div>}
                    {this.state.criticalConfig && (
                        <MonacoEditor
                            content={this.state.criticalConfig.Contents}
                            language="json"
                            onDidContentChange={this.onDidContentChange}
                            onDidSave={this.onDidSave}
                        />
                    )}
                </div>
                <button onClick={this.onDidSave}>Save changes</button>
            </div>
        )
    }

    private onDidContentChange = (content: string) => this.setState({ content })

    private onDidSave = () => {
        this.subscriptions.add(
            ajax({
                url: '/update',
                method: 'POST',
                body: JSON.stringify({
                    LastID: this.state.criticalConfig.ID,
                    Contents: this.state.content,
                }),
            }).subscribe(resp => {
                if (resp.status !== 200) {
                    console.error(resp.status + ' ' + resp.responseText)
                }
                alert('Saved successfully!')
            })
        )
    }
}
