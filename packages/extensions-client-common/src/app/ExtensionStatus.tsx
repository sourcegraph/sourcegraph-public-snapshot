import * as H from 'history'
import * as React from 'react'
import { Subject, Subscription } from 'rxjs'
import { distinctUntilChanged, map, switchMap } from 'rxjs/operators'
import { ExtensionConnection } from 'sourcegraph/module/client/controller'
import { BrowserConsoleTracer, Trace } from 'sourcegraph/module/protocol/jsonrpc2/trace'
import { ControllerProps } from '../client/controller'
import { ConfigurationSubject, Settings } from '../settings'
import { PopoverButton } from '../ui/generic/PopoverButton'
import { Toggle } from '../ui/generic/Toggle'

interface Props<S extends ConfigurationSubject, C extends Settings> extends ControllerProps<S, C> {
    caretIcon: React.ComponentType<{
        className: 'icon-inline' | string
        onClick?: () => void
    }>

    loaderIcon: React.ComponentType<{
        className: 'icon-inline' | string
        onClick?: () => void
    }>
    link: React.ComponentType<{ id: string }>
}

interface State {
    /** The extension clients, or undefined while loading. */
    extensions?: ExtensionConnection[]

    /** Whether to log traces of communication with extensions. */
    trace?: boolean
}

export class ExtensionStatus<S extends ConfigurationSubject, C extends Settings> extends React.PureComponent<
    Props<S, C>,
    State
> {
    public static TRACE_STORAGE_KEY = 'traceExtensions'

    public state: State = { trace: localStorage.getItem(ExtensionStatus.TRACE_STORAGE_KEY) !== null }

    private componentUpdates = new Subject<Props<S, C>>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        const extensionsController = this.componentUpdates.pipe(
            map(({ extensionsController }) => extensionsController),
            distinctUntilChanged()
        )

        this.subscriptions.add(
            extensionsController
                .pipe(
                    switchMap(extensionsController => extensionsController.clientEntries),
                    map(extensions => ({ extensions }))
                )
                .subscribe(stateUpdate => this.setState(stateUpdate), err => console.error(err))
        )

        this.componentUpdates.next(this.props)
    }

    public componentDidUpdate(): void {
        this.componentUpdates.next(this.props)
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        return (
            <div className="extension-status card border-0">
                <div className="card-header">Active extensions (DEBUG)</div>
                {this.state.extensions ? (
                    this.state.extensions.length > 0 ? (
                        <div className="list-group list-group-flush">
                            {this.state.extensions.map(({ key }, i) => (
                                <div
                                    key={i}
                                    className="list-group-item py-2 d-flex align-items-center justify-content-between"
                                >
                                    <this.props.link id={key.id} />
                                </div>
                            ))}
                        </div>
                    ) : (
                        <span className="card-body">No active extensions.</span>
                    )
                ) : (
                    <span className="card-body">
                        <this.props.loaderIcon className="icon-inline" /> Loading extensions...
                    </span>
                )}
                <div className="card-body border-top d-flex justify-content-end align-items-center">
                    <label htmlFor="extension-status__trace" className="mr-2 mb-0">
                        Log to devtools console{' '}
                    </label>
                    <Toggle
                        id="extension-status__trace"
                        onToggle={this.onToggleTrace}
                        value={this.state.trace}
                        title="Toggle extension trace logging to devtools console"
                    />
                </div>
            </div>
        )
    }

    private onToggleTrace = () => {
        this.setState(
            prevState => ({ trace: !prevState.trace }),
            () => {
                if (this.state.trace) {
                    localStorage.setItem(ExtensionStatus.TRACE_STORAGE_KEY, 'true')
                } else {
                    localStorage.removeItem(ExtensionStatus.TRACE_STORAGE_KEY)
                }

                // Update trace setting for all existing connections.
                if (this.state.extensions) {
                    for (const e of this.state.extensions) {
                        e.connection
                            .then(c =>
                                c.trace(
                                    this.state.trace ? Trace.Verbose : Trace.Off,
                                    new BrowserConsoleTracer(e.key.id)
                                )
                            )
                            .catch(err => console.error(err))
                    }
                }
            }
        )
    }
}

/** A button that toggles the visibility of the ExtensionStatus element in a popover. */
export class ExtensionStatusPopover<S extends ConfigurationSubject, C extends Settings> extends React.PureComponent<
    Props<S, C> & { location: H.Location }
> {
    public render(): JSX.Element | null {
        return (
            <PopoverButton
                caretIcon={this.props.caretIcon}
                placement="auto-end"
                hideOnChange={this.props.location}
                popoverElement={<ExtensionStatus {...this.props} />}
            >
                <span className="text-muted">Ext</span>
            </PopoverButton>
        )
    }
}
