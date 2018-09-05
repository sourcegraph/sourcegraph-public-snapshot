import { Client, ClientState } from 'sourcegraph/module/client/client'
import { ClientKey } from 'sourcegraph/module/environment/controller'
import { Trace } from 'sourcegraph/module/jsonrpc2/trace'
import * as React from 'react'
import { combineLatest, of, Subject, Subscription } from 'rxjs'
import { distinctUntilChanged, map, switchMap } from 'rxjs/operators'
import { updateSavedClientTrace } from '../client/client'
import { ControllerProps } from '../client/controller'
import { ConfigurationSubject, Settings } from '../settings'
import { PopoverButton } from '../ui/generic/PopoverButton'

interface Props<S extends ConfigurationSubject, C extends Settings> extends ControllerProps<S, C> {
    caretIcon: React.ComponentType<{
        className: 'icon-inline' | string
        onClick?: () => void
    }>

    loaderIcon: React.ComponentType<{
        className: 'icon-inline' | string
        onClick?: () => void
    }>
}

interface State {
    /** The extension clients, or undefined while loading. */
    clients?: { client: Client; key: ClientKey; state: ClientState }[]
}

export class ExtensionStatus<S extends ConfigurationSubject, C extends Settings> extends React.PureComponent<
    Props<S, C>,
    State
> {
    public state: State = {}

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
                    switchMap(extensionsController =>
                        extensionsController.clientEntries.pipe(
                            switchMap(
                                clientEntries =>
                                    clientEntries.length === 0
                                        ? of([])
                                        : combineLatest(
                                              clientEntries.map(({ client, key }) =>
                                                  client.state.pipe(
                                                      distinctUntilChanged(),
                                                      map(state => ({ state, client, key }))
                                                  )
                                              )
                                          )
                            )
                        )
                    ),
                    map(clients => ({ clients }))
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
                <div className="card-header">Extensions</div>
                {this.state.clients ? (
                    this.state.clients.length > 0 ? (
                        <div className="list-group list-group-flush">
                            {this.state.clients.map(({ client, key, state }, i) => (
                                <div
                                    key={i}
                                    className="extension-status__client list-group-item d-flex align-items-center justify-content-between py-2"
                                >
                                    <span className="d-flex align-items-center">
                                        <span data-tooltip={key.root || 'no root'}>{client.id}</span>
                                        <span className={`badge badge-${clientStateBadgeClass(state)} ml-1`}>
                                            {ClientState[state]}
                                        </span>
                                    </span>
                                    <div className="extension-status__client-actions d-flex align-items-center ml-3">
                                        <button
                                            className={`btn btn-sm btn-${
                                                client.trace === Trace.Off ? 'outline-' : ''
                                            }info py-0 px-1`}
                                            // tslint:disable-next-line:jsx-no-lambda
                                            onClick={() => this.onClientTraceClick(client, key)}
                                            data-tooltip={`${
                                                client.trace === Trace.Off ? 'Enable' : 'Disable'
                                            } trace in console`}
                                        >
                                            Log
                                        </button>
                                        {client.needsStop() && (
                                            <button
                                                className="btn btn-sm btn-outline-danger py-0 px-1 ml-1"
                                                // tslint:disable-next-line:jsx-no-lambda
                                                onClick={() => this.onClientStopClick(client)}
                                            >
                                                Stop
                                            </button>
                                        )}
                                        {!client.needsStop() && (
                                            <button
                                                className="btn btn-sm btn-outline-success py-0 px-1 ml-1"
                                                // tslint:disable-next-line:jsx-no-lambda
                                                onClick={() => this.onClientActivateClick(client)}
                                            >
                                                Start
                                            </button>
                                        )}
                                        {client.needsStop() && (
                                            <button
                                                className="btn btn-sm btn-outline-warning py-0 px-1 ml-1"
                                                // tslint:disable-next-line:jsx-no-lambda
                                                onClick={() => this.onClientResetClick(client)}
                                            >
                                                Restart
                                            </button>
                                        )}
                                    </div>
                                </div>
                            ))}
                        </div>
                    ) : (
                        <span className="card-body">No clients.</span>
                    )
                ) : (
                    <span className="card-body">
                        <this.props.loaderIcon className="icon-inline" /> Loading clients...
                    </span>
                )}
            </div>
        )
    }

    private onClientTraceClick = (client: Client, key: ClientKey) => {
        client.trace = client.trace === Trace.Verbose ? Trace.Off : Trace.Verbose
        updateSavedClientTrace(key, client.trace)
        this.forceUpdate()
    }

    private onClientStopClick = (client: Client) => client.stop()

    private onClientActivateClick = (client: Client) => client.activate()

    private onClientResetClick = (client: Client) => {
        let p = Promise.resolve<void>(void 0)
        if (client.needsStop()) {
            p = client.stop()
        }
        p.then(() => client.activate(), err => console.error(err))
    }
}

function clientStateBadgeClass(state: ClientState): string {
    switch (state) {
        case ClientState.Initial:
            return 'secondary'
        case ClientState.Connecting:
            return 'info'
        case ClientState.Initializing:
            return 'info'
        case ClientState.ActivateFailed:
            return 'danger'
        case ClientState.Active:
            return 'success'
        case ClientState.ShuttingDown:
            return 'warning'
        case ClientState.Stopped:
            return 'danger'
    }
}

/** A button that toggles the visibility of the ExtensionStatus element in a popover. */
export class ExtensionStatusPopover<S extends ConfigurationSubject, C extends Settings> extends React.PureComponent<
    Props<S, C>
> {
    public render(): JSX.Element | null {
        return (
            <PopoverButton
                caretIcon={this.props.caretIcon}
                placement="auto-end"
                globalKeyBinding="X"
                popoverElement={<ExtensionStatus {...this.props} />}
            >
                <span className="text-muted">Ext</span>
            </PopoverButton>
        )
    }
}
