import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import * as H from 'history'
import * as React from 'react'
import { Subject, Subscription } from 'rxjs'
import { distinctUntilChanged, map, switchMap } from 'rxjs/operators'
import { ExecutableExtension } from '../api/client/services/extensionsService'
import { PopoverButton } from '../components/PopoverButton'
import { Toggle } from '../components/Toggle'
import { ExtensionsControllerProps } from '../extensions/controller'
import { PlatformContextProps } from '../platform/context'

interface Props extends ExtensionsControllerProps, PlatformContextProps {
    link: React.ComponentType<{ id: string }>
}

interface State {
    /** The extension IDs of extensions that are active, or undefined while loading. */
    extensions?: Pick<ExecutableExtension, 'id'>[]

    /** Whether to log traces of communication with extensions. */
    traceExtensionHostCommunication?: boolean
}

export class ExtensionStatus extends React.PureComponent<Props, State> {
    public state: State = {}

    private componentUpdates = new Subject<Props>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        const extensionsController = this.componentUpdates.pipe(
            map(({ extensionsController }) => extensionsController),
            distinctUntilChanged()
        )
        this.subscriptions.add(
            extensionsController
                .pipe(
                    switchMap(extensionsController => extensionsController.services.extensions.activeExtensions),
                    map(extensions => ({ extensions }))
                )
                .subscribe(stateUpdate => this.setState(stateUpdate), err => console.error(err))
        )

        const platformContext = this.componentUpdates.pipe(
            map(({ platformContext }) => platformContext),
            distinctUntilChanged()
        )
        this.subscriptions.add(
            platformContext
                .pipe(
                    switchMap(({ traceExtensionHostCommunication }) => traceExtensionHostCommunication),
                    map(traceExtensionHostCommunication => ({ traceExtensionHostCommunication }))
                )
                .subscribe(stateUpdate => this.setState(stateUpdate))
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
                            {this.state.extensions.map(({ id }, i) => (
                                <div
                                    key={i}
                                    className="list-group-item py-2 d-flex align-items-center justify-content-between"
                                >
                                    <this.props.link id={id} />
                                </div>
                            ))}
                        </div>
                    ) : (
                        <span className="card-body">No active extensions.</span>
                    )
                ) : (
                    <span className="card-body">
                        <LoadingSpinner className="icon-inline" /> Loading extensions...
                    </span>
                )}
                <div className="card-body border-top d-flex justify-content-end align-items-center">
                    <label htmlFor="extension-status__trace" className="mr-2 mb-0">
                        Log to devtools console{' '}
                    </label>
                    <Toggle
                        id="extension-status__trace"
                        onToggle={this.onToggleTrace}
                        value={this.state.traceExtensionHostCommunication}
                        title="Toggle extension trace logging to devtools console"
                    />
                </div>
            </div>
        )
    }

    private onToggleTrace = () => {
        this.props.platformContext.traceExtensionHostCommunication.next(!this.state.traceExtensionHostCommunication)
    }
}

/** A button that toggles the visibility of the ExtensionStatus element in a popover. */
export class ExtensionStatusPopover extends React.PureComponent<Props & { location: H.Location }> {
    public render(): JSX.Element | null {
        return (
            <PopoverButton
                placement="auto-end"
                hideOnChange={this.props.location}
                popoverElement={<ExtensionStatus {...this.props} />}
            >
                <span className="text-muted">Ext</span>
            </PopoverButton>
        )
    }
}
