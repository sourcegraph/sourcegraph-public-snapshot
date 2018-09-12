import * as React from 'react'
import { combineLatest, of, Subject, Subscription } from 'rxjs'
import { distinctUntilChanged, map, switchMap } from 'rxjs/operators'
import { ExtensionConnectionKey } from 'sourcegraph/module/client/controller'
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
    clients?: { key: ExtensionConnectionKey }[]
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
                                        : combineLatest(clientEntries.map(({ key }) => ({ key })))
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
                            {this.state.clients.map(({ key }, i) => (
                                <div
                                    key={i}
                                    className="extension-status__client list-group-item d-flex align-items-center justify-content-between py-2"
                                >
                                    <span className="d-flex align-items-center">{key.id}</span>
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
