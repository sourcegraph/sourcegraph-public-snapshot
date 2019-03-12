import { upperFirst } from 'lodash'
import * as React from 'react'
import { merge, Subject, Subscription } from 'rxjs'
import { debounceTime, takeUntil } from 'rxjs/operators'

export enum ConnectionErrors {
    AuthError,
    UnableToConnect,
}

interface StatusClassNames {
    connecting: 'warning'
    connected: 'success'
    error: 'error'
}

const statusClassNames: StatusClassNames = {
    connecting: 'warning',
    connected: 'success',
    error: 'error',
}

/**
 * This is the [Word-Joiner](https://en.wikipedia.org/wiki/Word_joiner) character.
 * We are using this as a &nbsp; that has no width to maintain line height when the
 * url is being updated (therefore no text is in the status indicator).
 */
const zeroWidthNbsp = '\u2060'

export interface ServerURLFormProps {
    className?: string
    status: keyof StatusClassNames
    connectionError?: ConnectionErrors

    value: string
    onChange: (value: string) => void
    onSubmit: () => void
    urlHasPermissions: boolean
    requestPermissions: () => void

    /**
     * Overrides `this.props.status` and `this.state.isUpdating` in order to
     * display the `isUpdating` UI state. This is only intended for use in storybooks.
     */
    overrideUpdatingState?: boolean
}

interface State {
    isUpdating: boolean
}

export class ServerURLForm extends React.Component<ServerURLFormProps> {
    public state: State = { isUpdating: false }

    private inputElement = React.createRef<HTMLInputElement>()

    private componentUpdates = new Subject<State>()
    private changes = new Subject<string>()
    private submits = new Subject<void>()

    private subscriptions = new Subscription()

    constructor(props: ServerURLFormProps) {
        super(props)

        this.subscriptions.add(
            this.changes.subscribe(value => {
                this.props.onChange(value)
                this.setState({ isUpdating: true })
            })
        )

        const submitAfterInactivity = this.changes.pipe(
            debounceTime(5000),
            takeUntil(this.submits)
        )

        this.subscriptions.add(
            merge(this.submits, submitAfterInactivity).subscribe(() => {
                this.props.onSubmit()
                this.setState({ isUpdating: false })
            })
        )
    }

    public componentDidUpdate(): void {
        this.componentUpdates.next(this.state)
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): React.ReactNode {
        return (
            // tslint:disable-next-line:jsx-ban-elements
            <form className={`server-url-form ${this.props.className || ''}`} onSubmit={this.handleSubmit}>
                <label className="server-url-form__label">Sourcegraph URL</label>
                <div className="server-url-form__input-container">
                    <div className="server-url-form__input-container__status">
                        <span
                            className={
                                'server-url-form__input-container__status__indicator ' +
                                'server-url-form__input-container__status__indicator--' +
                                (this.isUpdating ? 'default' : statusClassNames[this.props.status])
                            }
                        />
                        <span className="server-url-form__input-container__status__text">
                            {this.isUpdating ? zeroWidthNbsp : upperFirst(this.props.status)}
                        </span>
                    </div>
                    <input
                        type="text"
                        ref={this.inputElement}
                        value={this.props.value}
                        className="server-url-form__input-container__input"
                        onChange={this.handleChange}
                    />
                </div>
                {!this.state.isUpdating && this.props.connectionError === ConnectionErrors.AuthError && (
                    <div className="server-url-form__error">
                        Authentication to Sourcegraph failed.{' '}
                        <a href={this.props.value} target="_blank">
                            Sign in to your instance
                        </a>{' '}
                        to continue.
                    </div>
                )}
                {!this.state.isUpdating && this.props.connectionError === ConnectionErrors.UnableToConnect && (
                    <div className="server-url-form__error">
                        <p>
                            Unable to connect to{' '}
                            <a href={this.props.value} target="_blank">
                                {this.props.value}
                            </a>
                            . Ensure the URL is correct and you are{' '}
                            <a href={this.props.value + '/sign-in'} target="_blank">
                                logged in
                            </a>
                            .
                        </p>
                        {!this.props.urlHasPermissions && (
                            <p>
                                You may need to{' '}
                                <a href="#" onClick={this.props.requestPermissions}>
                                    grant the Sourcegraph browser extension additional permissions
                                </a>{' '}
                                for this URL.
                            </p>
                        )}
                        <p>
                            <b>Site admins:</b> ensure that{' '}
                            <a
                                href="https://docs.sourcegraph.com/admin/site_config/all#auth-accesstokens-object"
                                target="_blank"
                            >
                                all users can create access tokens
                            </a>
                            .
                        </p>
                    </div>
                )}
            </form>
        )
    }

    private handleChange = ({ target: { value } }: React.ChangeEvent<HTMLInputElement>) => {
        this.changes.next(value)
    }

    private handleSubmit = (event: React.FormEvent<HTMLFormElement>) => {
        event.preventDefault()

        this.submits.next()
    }

    private get isUpdating(): boolean {
        if (typeof this.props.overrideUpdatingState !== 'undefined') {
            console.warn(
                '<ServerURLForm /> - You are using the `overrideUpdatingState` prop which is ' +
                    'only intended for use with storybooks. Keeping this state in multiple places can ' +
                    'lead to race conditions and will be hard to maintain.'
            )

            return this.props.overrideUpdatingState
        }

        return this.state.isUpdating
    }
}
