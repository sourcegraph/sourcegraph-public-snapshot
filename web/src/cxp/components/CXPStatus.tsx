import BrowserStopIcon from '@sourcegraph/icons/lib/BrowserStop'
import InfoIcon from '@sourcegraph/icons/lib/Info'
import Loader from '@sourcegraph/icons/lib/Loader'
import RefreshIcon from '@sourcegraph/icons/lib/Refresh'
import { Client as CXPClient, ClientState as CXPClientState } from 'cxp/lib/client/client'
import { Trace } from 'cxp/lib/jsonrpc2/trace'
import * as React from 'react'
import { combineLatest, of, Subject, Subscription } from 'rxjs'
import { distinctUntilChanged, map, switchMap } from 'rxjs/operators'
import { PopoverButton } from '../../components/PopoverButton'
import { CXPControllerProps, CXPEnvironmentProps } from '../CXPEnvironment'

interface Props extends CXPEnvironmentProps, CXPControllerProps {}

interface State {
    /** The CXP clients, or undefined while loading. */
    clients?: { client: CXPClient; state: CXPClientState }[]
}

export class CXPStatus extends React.PureComponent<Props, State> {
    public state: State = {}

    private componentUpdates = new Subject<Props>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        const cxpController = this.componentUpdates.pipe(
            map(({ cxpController }) => cxpController),
            distinctUntilChanged()
        )

        this.subscriptions.add(
            cxpController
                .pipe(
                    switchMap(cxpController =>
                        cxpController.clients.pipe(
                            switchMap(
                                clients =>
                                    clients.length === 0
                                        ? of([])
                                        : combineLatest(
                                              clients.map(client =>
                                                  client.state.pipe(
                                                      distinctUntilChanged(),
                                                      map(state => ({ state, client }))
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
            <div className="cxp-status card border-0">
                <div className="card-header">CXP clients</div>
                {this.state.clients ? (
                    this.state.clients.length > 0 ? (
                        <div className="list-group list-group-flush">
                            {this.state.clients.map(({ client, state }, i) => (
                                <div
                                    key={i}
                                    className="cxp-status__client list-group-item d-flex align-items-center justify-content-between py-2"
                                >
                                    <span className="d-flex align-items-center">
                                        {client.id}
                                        <span className={`badge badge-${clientStateBadgeClass(state)} ml-1`}>
                                            {CXPClientState[state]}
                                        </span>
                                    </span>
                                    <div className="cxp-status__client-actions d-flex align-items-center ml-3">
                                        <button
                                            className={`btn btn-sm btn-${
                                                client.trace === Trace.Off ? 'outline-' : ''
                                            }info p-0`}
                                            // tslint:disable-next-line:jsx-no-lambda
                                            onClick={() => this.onClientTraceClick(client)}
                                            data-tooltip={`${
                                                client.trace === Trace.Off ? 'Enable' : 'Disable'
                                            } trace logging in devtools console`}
                                        >
                                            <InfoIcon className="icon-inline" />
                                        </button>
                                        {client.needsStop() && (
                                            <button
                                                className="btn btn-sm btn-outline-danger p-0 ml-1"
                                                // tslint:disable-next-line:jsx-no-lambda
                                                onClick={() => this.onClientStopClick(client)}
                                                data-tooltip="Stop client"
                                            >
                                                <BrowserStopIcon className="icon-inline" />
                                            </button>
                                        )}
                                        {!client.needsStop() && (
                                            <button
                                                className="btn btn-sm btn-outline-success p-0 ml-1"
                                                // tslint:disable-next-line:jsx-no-lambda
                                                onClick={() => this.onClientStartClick(client)}
                                                data-tooltip="Start client"
                                            >
                                                <RefreshIcon className="icon-inline" />
                                            </button>
                                        )}
                                        {client.needsStop() && (
                                            <button
                                                className="btn btn-sm btn-outline-warning p-0 ml-1"
                                                // tslint:disable-next-line:jsx-no-lambda
                                                onClick={() => this.onClientRestartClick(client)}
                                                data-tooltip="Restart client"
                                            >
                                                <RefreshIcon className="icon-inline" />
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
                        <Loader className="icon-inline" /> Loading clients...
                    </span>
                )}
            </div>
        )
    }

    private onClientTraceClick = (client: CXPClient) => {
        client.trace = client.trace === Trace.Verbose ? Trace.Off : Trace.Verbose
        this.forceUpdate()
    }

    private onClientStopClick = (client: CXPClient) => client.stop()

    private onClientStartClick = (client: CXPClient) => client.start()

    private onClientRestartClick = (client: CXPClient) => {
        let p = Promise.resolve<void>(void 0)
        if (client.needsStop()) {
            p = client.stop()
        }
        p.then(() => client.start(), err => console.error(err))
    }
}

function clientStateBadgeClass(state: CXPClientState): string {
    switch (state) {
        case CXPClientState.Initial:
            return 'secondary'
        case CXPClientState.Starting:
            return 'info'
        case CXPClientState.StartFailed:
            return 'danger'
        case CXPClientState.Initializing:
            return 'info'
        case CXPClientState.Active:
            return 'success'
        case CXPClientState.Stopping:
            return 'warning'
        case CXPClientState.Stopped:
            return 'danger'
    }
}

/** A button that toggles the visibility of the CXPStatus element in a popover. */
export const CXPStatusPopover: React.SFC<Props> = props => (
    <PopoverButton placement="auto-end" popoverElement={<CXPStatus {...props} />}>
        CXP
    </PopoverButton>
)
