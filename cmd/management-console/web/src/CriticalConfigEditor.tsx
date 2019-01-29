import * as React from 'react'
import { from, interval, Observable, of, Subject, Subscription, timer } from 'rxjs'
import { ajax } from 'rxjs/ajax'
import { delay, distinctUntilChanged, mapTo, startWith, takeUntil, catchError } from 'rxjs/operators'
import './CriticalConfigEditor.scss'
import { MonacoEditor } from './MonacoEditor'

const DEBUG_LOADING_STATE_DELAY = 0 // ms

/**
 * Amount of time to wait before showing the loading indicator.
 */
const WAIT_BEFORE_SHOWING_LOADER = 250 // ms

// TODO(slimsag): future: Warn user if they are discarding changes
// TODO(slimsag): future: Explicit discard changes button?
// TODO(slimsag): future: Better button styling
// TODO(slimsag): future: Better link styling
// TODO(slimsag): future: Better 'loading' state styling

/**
 * The success response from the API /get and /update endpoints.
 */
interface Configuration {
    /**
     * The unique ID of this configuration version.
     */
    ID: string

    /**
     * The literal JSONC configuration.
     */
    Contents: string
}

interface ConfigurationContents {
    /**
     * The instance's license key.
     */
    LicenseKey: string
}

interface LicenseKeyInfo {
    /**
     * The number of users on an instance.
     */
    UserCount: number
    ExpiresAt: string
}

/**
 * The parameters that mut be POST to the /update endpoint.
 */
interface UpdateParams {
    /**
     * The last Configuration.ID value the client was aware of. If outdated,
     * the update will fail.
     */
    LastID: string

    /**
     * The literal JSONC configuration.
     */
    Contents: string
}

interface Props {}

interface State {
    /** The current config content according to the server. */
    criticalConfig: Configuration | null

    /** The current content in the editor. */
    content: string | null

    /** Whether or not the loader can be shown yet, iff criticalConfig === null */
    canShowLoader: boolean

    /** The instance's license key, as specified in the config. */
    licenseKey: string | null

    /** Whether or not to show a "Saving..." indicator */
    showSaving: boolean

    /** Whether or not to show a "Saved!" indicator */
    showSaved: boolean

    /** Whether or not to show a saving error indicator */
    showSavingError: string | null

    userCount: number
    expiresAt: string | null
}

export class CriticalConfigEditor extends React.Component<Props, State> {
    public state: State = {
        criticalConfig: null,
        content: null,
        canShowLoader: false,
        expiresAt: null,
        licenseKey: null,
        showSaving: false,
        showSaved: false,
        showSavingError: null,
        userCount: 0,
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
            ajax('/api/get')
                .pipe(
                    delay(DEBUG_LOADING_STATE_DELAY),
                    catchError(err => of(err.xhr))
                )
                .subscribe(resp => {
                    if (resp.status !== 200) {
                        const msg = 'error saving: ' + resp.status
                        console.error(msg)
                        alert(msg) // TODO(slimsag): Better general error state here.
                        return
                    }

                    const config = resp.response as Configuration
                    this.setState({
                        criticalConfig: config,
                        content: config.Contents,
                    })
                })
        )

        // Load the initial critical config.
        this.subscriptions.add(
            ajax('/api/license')
                .pipe(
                    delay(DEBUG_LOADING_STATE_DELAY),
                    catchError(err => of(err.xhr))
                )
                .subscribe(resp => {
                    if (resp.status !== 200) {
                        const msg = 'error saving: ' + resp.status
                        console.error(msg)
                        alert(msg) // TODO(slimsag): Better general error state here.
                        return
                    }

                    const license = resp.response as LicenseKeyInfo
                    this.setState({
                        userCount: license.UserCount,
                        expiresAt: license.ExpiresAt,
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
                    <div>{this.state.userCount}</div>
                    <div>{this.state.expiresAt}</div>
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
                {this.state.showSaving && <span className="critical-config-editor__status-indicator">Saving...</span>}
                {this.state.showSaved && (
                    <span className="critical-config-editor__status-indicator critical-config-editor__status-indicator--success">
                        Saved!
                    </span>
                )}
                {this.state.showSavingError && (
                    <span className="critical-config-editor__status-indicator critical-config-editor__status-indicator--error">
                        {this.state.showSavingError}
                    </span>
                )}
            </div>
        )
    }

    private onDidContentChange = (content: string) => this.setState({ content })

    private onDidSave = () => {
        this.setState(
            {
                showSaving: true,
                showSaved: false,
                showSavingError: null,
            },
            () =>
                this.subscriptions.add(
                    ajax({
                        url: '/api/update',
                        method: 'POST',
                        body: JSON.stringify({
                            LastID: this.state.criticalConfig.ID,
                            Contents: this.state.content,
                        } as UpdateParams),
                    })
                        .pipe(catchError(err => of(err.xhr)))
                        .subscribe(resp => {
                            if (resp.status !== 200) {
                                const msg =
                                    resp.status === 409
                                        ? 'error: someone else has already applied a newer edit'
                                        : 'error: ' + resp.status
                                console.error(msg)
                                this.setState({
                                    showSaving: false,
                                    showSaved: false,
                                    showSavingError: msg,
                                })
                                return
                            }
                            const config = resp.response as Configuration
                            this.setState({
                                criticalConfig: config,
                                content: config.Contents,
                                showSaving: false,
                                showSaved: true,
                                showSavingError: null,
                            })

                            // Hide the saved indicator after 2.5s.
                            setTimeout(() => this.setState({ showSaved: false }), 2500)
                        })
                )
        )
    }
}
