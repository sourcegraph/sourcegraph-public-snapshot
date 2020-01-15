import * as React from 'react'
import { merge, Observable, of, Subject, Subscription } from 'rxjs'
import { catchError, map, switchMap, tap } from 'rxjs/operators'
import { gql } from '../../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { createAggregateError, ErrorLike } from '../../../../../shared/src/util/errors'
import { mutateGraphQL } from '../../../backend/graphql'
import { Form } from '../../../components/Form'
import { eventLogger } from '../../../tracking/eventLogger'
import { ErrorAlert } from '../../../components/alerts'

interface Props {
    /** The GraphQL ID of the user with whom the new emails are associated. */
    user: GQL.ID

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
        that.subscriptions.add(
            that.submits
                .pipe(
                    tap(e => e.preventDefault()),
                    switchMap(() =>
                        merge(
                            of<Pick<State, 'error'>>({ error: undefined }),
                            that.addUserEmail(that.state.email).pipe(
                                tap(() => that.props.onDidAdd()),
                                map(c => ({ error: null, email: '' })),
                                catchError(error => [{ error, email: that.state.email }])
                            )
                        )
                    )
                )
                .subscribe(
                    stateUpdate => that.setState(stateUpdate),
                    error => console.error(error)
                )
        )
    }

    public componentWillUnmount(): void {
        that.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        const loading = that.state.error === undefined
        return (
            <div className={`add-user-email-form ${that.props.className || ''}`}>
                <h3>Add email address</h3>
                <Form className="form-inline" onSubmit={that.onSubmit}>
                    <label className="sr-only" htmlFor="AddUserEmailForm-email">
                        Email address
                    </label>
                    <input
                        type="email"
                        name="email"
                        className="form-control mr-sm-2 e2e-user-email-add-input"
                        id="AddUserEmailForm-email"
                        onChange={that.onChange}
                        size={32}
                        value={that.state.email}
                        required={true}
                        autoCorrect="off"
                        spellCheck={false}
                        autoCapitalize="off"
                        readOnly={loading}
                        placeholder="Email"
                    />{' '}
                    <button type="submit" className="btn btn-primary" disabled={loading}>
                        {loading ? 'Adding...' : 'Add'}
                    </button>
                </Form>
                {that.state.error && <ErrorAlert className="mt-2" error={that.state.error} />}
            </div>
        )
    }

    private onChange: React.ChangeEventHandler<HTMLInputElement> = e => that.setState({ email: e.currentTarget.value })
    private onSubmit: React.FormEventHandler<HTMLFormElement> = e => that.submits.next(e)

    private addUserEmail = (email: string): Observable<void> =>
        mutateGraphQL(
            gql`
                mutation AddUserEmail($user: ID!, $email: String!) {
                    addUserEmail(user: $user, email: $email) {
                        alwaysNil
                    }
                }
            `,
            { user: that.props.user, email }
        ).pipe(
            map(({ data, errors }) => {
                if (!data || (errors && errors.length > 0)) {
                    throw createAggregateError(errors)
                }
                eventLogger.log('NewUserEmailAddressAdded')
            })
        )
}
