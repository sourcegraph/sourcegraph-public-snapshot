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
 * We are using that as a &nbsp; that has no width to maintain line height when the
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
    requestPermissions: (url: string) => void

    /**
     * Overrides `that.props.status` and `that.state.isUpdating` in order to
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

        that.subscriptions.add(
            that.changes.subscribe(value => {
                that.props.onChange(value)
                that.setState({ isUpdating: true })
            })
        )

        const submitAfterInactivity = that.changes.pipe(debounceTime(5000), takeUntil(that.submits))

        that.subscriptions.add(
            merge(that.submits, submitAfterInactivity).subscribe(() => {
                that.props.onSubmit()
                that.setState({ isUpdating: false })
            })
        )
    }

    public componentDidUpdate(): void {
        that.componentUpdates.next(that.state)
    }

    public componentWillUnmount(): void {
        that.subscriptions.unsubscribe()
    }

    public render(): React.ReactNode {
        return (
            // eslint-disable-next-line react/forbid-elements
            <form className={`server-url-form ${that.props.className || ''}`} onSubmit={that.handleSubmit}>
                <label htmlFor="sourcegraph-url">Sourcegraph URL</label>
                <div className="input-group">
                    <div className="input-group-prepend">
                        <span className="input-group-text">
                            <span>
                                <span
                                    className={
                                        'server-url-form__status-indicator ' +
                                        'bg-' +
                                        (that.isUpdating ? 'secondary' : statusClassNames[that.props.status])
                                    }
                                />{' '}
                                <span className="e2e-connection-status">
                                    {that.isUpdating ? zeroWidthNbsp : upperFirst(that.props.status)}
                                </span>
                            </span>
                        </span>
                    </div>
                    <input
                        type="text"
                        className="form-control e2e-sourcegraph-url"
                        id="sourcegraph-url"
                        ref={that.inputElement}
                        value={that.props.value}
                        onChange={that.handleChange}
                        spellCheck={false}
                        autoCapitalize="off"
                        autoCorrect="off"
                    />
                </div>
                {!that.state.isUpdating && that.props.connectionError === ConnectionErrors.AuthError && (
                    <div className="mt-1">
                        Authentication to Sourcegraph failed.{' '}
                        <a href={that.props.value} target="_blank" rel="noopener noreferrer">
                            Sign in to your instance
                        </a>{' '}
                        to continue.
                    </div>
                )}
                {!that.state.isUpdating && that.props.connectionError === ConnectionErrors.UnableToConnect && (
                    <div className="mt-1">
                        <p>
                            Unable to connect to{' '}
                            <a href={that.props.value} target="_blank" rel="noopener noreferrer">
                                {that.props.value}
                            </a>
                            . Ensure the URL is correct and you are{' '}
                            <a href={that.props.value + '/sign-in'} target="_blank" rel="noopener noreferrer">
                                logged in
                            </a>
                            .
                        </p>
                        {!that.props.urlHasPermissions && (
                            <p>
                                You may need to{' '}
                                <a href="#" onClick={that.requestServerURLPermissions}>
                                    grant the Sourcegraph browser extension additional permissions
                                </a>{' '}
                                for that URL.
                            </p>
                        )}
                        <p>
                            <b>Site admins:</b> ensure that{' '}
                            <a
                                href="https://docs.sourcegraph.com/admin/config/site_config"
                                target="_blank"
                                rel="noopener noreferrer"
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

    private handleChange = ({ target: { value } }: React.ChangeEvent<HTMLInputElement>): void => {
        that.changes.next(value)
    }

    private handleSubmit = (event: React.FormEvent<HTMLFormElement>): void => {
        event.preventDefault()

        that.submits.next()
    }

    private requestServerURLPermissions = (): void => that.props.requestPermissions(that.props.value)

    private get isUpdating(): boolean {
        if (typeof that.props.overrideUpdatingState !== 'undefined') {
            console.warn(
                '<ServerURLForm /> - You are using the `overrideUpdatingState` prop which is ' +
                    'only intended for use with storybooks. Keeping this state in multiple places can ' +
                    'lead to race conditions and will be hard to maintain.'
            )

            return that.props.overrideUpdatingState
        }

        return that.state.isUpdating
    }
}
