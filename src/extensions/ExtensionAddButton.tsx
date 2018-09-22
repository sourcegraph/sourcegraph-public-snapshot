import * as React from 'react'
import { from, Subject, Subscription } from 'rxjs'
import { catchError, map, mapTo, startWith, switchMap, tap } from 'rxjs/operators'
import { ExtensionsProps } from '../context'
import { asError, ErrorLike, isErrorLike } from '../errors'
import { ConfigurationSubject, ConfiguredSubjectOrError, Settings } from '../settings'
import { ConfiguredExtension } from './extension'

const LOADING: 'loading' = 'loading'

interface Props<S extends ConfigurationSubject, C extends Settings> extends ExtensionsProps<S, C> {
    /** The extension that this button adds. */
    extension: ConfiguredExtension

    /** The configuration subject that this button adds the extension for. */
    subject: ConfiguredSubjectOrError<ConfigurationSubject, Settings>

    disabled?: boolean

    className?: string

    /**
     * Called to confirm the primary action. If the callback returns false, the action is not
     * performed.
     */
    confirm?: () => boolean

    /** Called when the component performs an update that requires the parent component to refresh data. */
    onUpdate: () => void
}

interface State {
    /** The operation's status: null when done or not started, 'loading', or an error. */
    operationResultOrError: typeof LOADING | null | ErrorLike
}

/** An button to add an extension. */
export class ExtensionAddButton<S extends ConfigurationSubject, C extends Settings> extends React.PureComponent<
    Props<S, C>,
    State
> {
    public state: State = { operationResultOrError: null }

    private clicks = new Subject<void>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        this.subscriptions.add(
            this.clicks
                .pipe(
                    switchMap(() =>
                        from(this.addExtensionForSubject(this.props.extension, this.props.subject)).pipe(
                            mapTo(null),
                            catchError(error => [asError(error) as ErrorLike]),
                            map(c => ({ operationResultOrError: c } as State)),
                            tap(() => this.props.onUpdate()),
                            startWith<State>({ operationResultOrError: LOADING })
                        )
                    )
                )
                .subscribe(stateUpdate => this.setState(stateUpdate), error => console.error(error))
        )
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        return (
            <button
                className={`${this.props.className} d-flex align-items-center`}
                disabled={this.props.disabled || this.state.operationResultOrError === 'loading'}
                onClick={this.onClick}
            >
                {this.props.children}
                {isErrorLike(this.state.operationResultOrError) && (
                    <small className="text-danger ml-2" title={this.state.operationResultOrError.message}>
                        <this.props.extensions.context.icons.Warning className="icon-inline" /> Error
                    </small>
                )}
            </button>
        )
    }

    private onClick: React.MouseEventHandler<HTMLElement> = () => {
        if (!this.props.confirm || this.props.confirm()) {
            this.clicks.next()
        }
    }

    private addExtensionForSubject = (
        extension: ConfiguredExtension,
        subject: ConfiguredSubjectOrError<ConfigurationSubject, Settings>
    ) =>
        this.props.extensions.context.updateExtensionSettings(subject.subject.id, {
            extensionID: extension.id,
            enabled: true,
        })
}
