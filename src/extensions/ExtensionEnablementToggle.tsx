import * as React from 'react'
import { combineLatest, Subject, Subscription, from } from 'rxjs'
import { catchError, distinctUntilChanged, map, mapTo, startWith, switchMap, tap } from 'rxjs/operators'
import { ExtensionsProps } from '../context'
import { asError, ErrorLike, isErrorLike } from '../errors'
import { ConfigurationCascadeProps, ConfigurationSubject, Settings } from '../settings'
import { Toggle } from '../ui/generic/Toggle'
import { ConfiguredExtension, isExtensionEnabled } from './extension'

interface Props<S extends ConfigurationSubject, C extends Settings>
    extends ConfigurationCascadeProps<S, C>,
        ExtensionsProps<S, C> {
    extension: ConfiguredExtension

    /** The subject whose settings are edited when the user toggles enablement using this component. */
    subject: Pick<ConfigurationSubject, 'id' | 'viewerCanAdminister'>

    /**
     * Called when this component results in the extension's enablement state being changed.
     */
    onChange: (enabled: boolean) => void

    className?: string
    tabIndex?: number
}

const LOADING: 'loading' = 'loading'

interface State {
    /** The toggle operation's status: null when not started, true when done, 'loading', or an error. */
    toggleOrError: null | typeof LOADING | true | ErrorLike
}

/**
 * Enables and disables the extension and displays the enablement state.
 */
export class ExtensionEnablementToggle<S extends ConfigurationSubject, C extends Settings> extends React.PureComponent<
    Props<S, C>,
    State
> {
    public state: State = { toggleOrError: null }

    private componentUpdates = new Subject<Props<S, C>>()
    private toggles = new Subject<boolean>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        const extensionChanges = this.componentUpdates.pipe(
            map(({ extension }) => extension),
            distinctUntilChanged((a, b) => a.id === b.id)
        )

        const enablementChanges = combineLatest(
            extensionChanges,
            this.componentUpdates.pipe(
                map(({ configurationCascade }) => configurationCascade && configurationCascade.merged)
            )
        ).pipe(map(([extension, settings]) => isExtensionEnabled(settings, extension.id)))

        // Reset toggleOrError compensation for stale enablement after we receive the new post-update value.
        this.subscriptions.add(enablementChanges.subscribe(() => this.setState({ toggleOrError: null })))

        this.subscriptions.add(
            this.toggles
                .pipe(
                    switchMap(enabled =>
                        from(
                            this.props.extensions.context.updateExtensionSettings(this.props.subject.id, {
                                extensionID: this.props.extension.id,
                                enabled,
                            })
                        ).pipe(
                            mapTo(true),
                            catchError(error => [asError(error) as ErrorLike]),
                            map(c => ({ toggleOrError: c } as State)),
                            tap(() => {
                                if (this.props.onChange) {
                                    this.props.onChange(enabled)
                                }
                            }),
                            startWith<State>({ toggleOrError: LOADING })
                        )
                    )
                )
                .subscribe(stateUpdate => this.setState(stateUpdate), error => console.error(error))
        )
    }

    public componentWillReceiveProps(props: Props<S, C>): void {
        this.componentUpdates.next(props)
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        if (this.props.extension === null) {
            return null
        }

        // Invert extension enablement if we changed the value but the change hasn't yet been synced to the server.
        const unadjustedIsEnabled = isExtensionEnabled(this.props.configurationCascade.merged, this.props.extension.id)
        const isEnabled = this.state.toggleOrError === LOADING ? !unadjustedIsEnabled : unadjustedIsEnabled

        return (
            <div className="d-flex align-items-center">
                {isErrorLike(this.state.toggleOrError) && (
                    <span className="text-danger" title={this.state.toggleOrError.message}>
                        <this.props.extensions.context.icons.Warning className="icon-inline" />
                    </span>
                )}
                <Toggle
                    value={isEnabled}
                    title={isEnabled ? 'Enabled (slide to disable)' : 'Disabled (slide to enable)'}
                    disabled={!this.props.subject.viewerCanAdminister}
                    onToggle={this.onChange}
                    tabIndex={this.props.tabIndex}
                />
            </div>
        )
    }

    private onChange = (value: boolean) => {
        this.toggles.next(value)
    }
}
