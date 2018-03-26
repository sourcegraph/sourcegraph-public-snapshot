import upperFirst from 'lodash/upperFirst'
import * as React from 'react'
import { Observable } from 'rxjs/Observable'
import { merge } from 'rxjs/observable/merge'
import { of } from 'rxjs/observable/of'
import { catchError } from 'rxjs/operators/catchError'
import { map } from 'rxjs/operators/map'
import { publishReplay } from 'rxjs/operators/publishReplay'
import { refCount } from 'rxjs/operators/refCount'
import { switchMap } from 'rxjs/operators/switchMap'
import { tap } from 'rxjs/operators/tap'
import { Subject } from 'rxjs/Subject'
import { Subscription } from 'rxjs/Subscription'
import { gql, mutateGraphQL } from '../../backend/graphql'
import { createAggregateError, ErrorLike } from '../../util/errors'

interface Props {
    /** The GraphQL ID of the user with whom the new emails are associated. */
    user: GQLID

    /** Called after successfully adding an email to the user. */
    onDidAdd: () => void

    className?: string
}

interface State {
    email: string
    error?: ErrorLike | null
}

export class AddUserEmailForm extends React.PureComponent<Props, State> {
    public state: State = { email: '', error: null }

    private submits = new Subject<React.FormEvent<HTMLFormElement>>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        this.subscriptions.add(
            this.submits
                .pipe(
                    tap(e => e.preventDefault()),
                    switchMap(() =>
                        merge(
                            of({ error: undefined }),
                            this.addUserEmail(this.state.email).pipe(
                                tap(() => this.props.onDidAdd()),
                                map(c => ({ error: null, email: '' })),
                                catchError(error => [{ error, email: this.state.email }]),
                                publishReplay<Pick<State, 'email' | 'error'>>(),
                                refCount()
                            )
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
        const loading = this.state.error === undefined
        return (
            <div className={`add-user-email-form ${this.props.className || ''}`}>
                <h3>Add email address</h3>
                <form className="form-inline" onSubmit={this.onSubmit}>
                    <label className="sr-only" htmlFor="AddUserEmailForm-email">
                        Email address
                    </label>
                    <input
                        type="email"
                        name="email"
                        className="form-control mr-sm-2"
                        id="AddUserEmailForm-email"
                        onChange={this.onChange}
                        size={32}
                        value={this.state.email}
                        required={true}
                        autoCorrect="off"
                        spellCheck={false}
                        autoCapitalize="off"
                        readOnly={loading}
                    />{' '}
                    <button type="submit" className="btn btn-primary" disabled={loading}>
                        {loading ? 'Adding...' : 'Add'}
                    </button>
                </form>
                {this.state.error && (
                    <div className="alert alert-danger mt-2">{upperFirst(this.state.error.message)}</div>
                )}
            </div>
        )
    }

    private onChange: React.ChangeEventHandler<HTMLInputElement> = e => this.setState({ email: e.currentTarget.value })
    private onSubmit: React.FormEventHandler<HTMLFormElement> = e => this.submits.next(e)

    private addUserEmail = (email: string): Observable<void> =>
        mutateGraphQL(
            gql`
                mutation AddUserEmail($user: ID!, $email: String!) {
                    addUserEmail(user: $user, email: $email) {
                        alwaysNil
                    }
                }
            `,
            { user: this.props.user, email }
        ).pipe(
            map(({ data, errors }) => {
                if (!data || (errors && errors.length > 0)) {
                    throw createAggregateError(errors)
                }
            })
        )
}
