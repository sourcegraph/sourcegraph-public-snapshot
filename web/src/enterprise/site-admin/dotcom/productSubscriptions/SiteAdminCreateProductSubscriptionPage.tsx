import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { Observable, Subject, Subscription } from 'rxjs'
import { catchError, map, startWith, switchMap, tap } from 'rxjs/operators'
import { gql } from '../../../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { asError, createAggregateError, ErrorLike, isErrorLike } from '../../../../../../shared/src/util/errors'
import { pluralize } from '../../../../../../shared/src/util/strings'
import { mutateGraphQL, queryGraphQL } from '../../../../backend/graphql'
import { Form } from '../../../../components/Form'
import { PageTitle } from '../../../../components/PageTitle'
import { eventLogger } from '../../../../tracking/eventLogger'
import { AccountEmailAddresses } from '../../../dotcom/productSubscriptions/AccountEmailAddresses'
import { ErrorAlert } from '../../../../components/alerts'

interface Props extends RouteComponentProps<{}> {
    authenticatedUser: GQL.IUser
}

/** A customer account. */
interface Account extends Pick<GQL.IUser, 'id' | 'username' | 'displayName'> {
    emails: Pick<GQL.IUserEmail, 'email' | 'verified'>[]
}

const LOADING: 'loading' = 'loading'

interface State {
    accountID: GQL.ID

    /**
     * The list of all possible accounts.
     */
    accountsOrError: Account[] | typeof LOADING | ErrorLike

    /**
     * The result of creating the product subscription, or null when not pending or complete, or loading, or an
     * error.
     */
    creationOrError:
        | null
        | Pick<GQL.IProductSubscription, 'id' | 'name' | 'url' | 'urlForSiteAdmin'>
        | typeof LOADING
        | ErrorLike
}

/**
 * Creates a product subscription for an account based on information provided in the displayed form.
 *
 * For use on Sourcegraph.com by Sourcegraph teammates only.
 */
export class SiteAdminCreateProductSubscriptionPage extends React.Component<Props, State> {
    private get emptyState(): Pick<State, 'accountID' | 'creationOrError'> {
        return {
            accountID: this.props.authenticatedUser.id,
            creationOrError: null,
        }
    }

    public state: State = {
        ...this.emptyState,
        accountsOrError: LOADING,
    }

    private submits = new Subject<void>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        eventLogger.logViewEvent('SiteAdminCreateProductSubscription')

        this.subscriptions.add(
            queryAccounts()
                .pipe(
                    catchError(err => [asError(err)]),
                    startWith(LOADING),
                    map(c => ({ accountsOrError: c }))
                )
                .subscribe(stateUpdate => this.setState(stateUpdate))
        )

        this.subscriptions.add(
            this.submits
                .pipe(
                    switchMap(() =>
                        createProductSubscription({ accountID: this.state.accountID }).pipe(
                            tap(({ url, urlForSiteAdmin }) => this.props.history.push(urlForSiteAdmin || url)),
                            catchError(err => [asError(err)]),
                            startWith(LOADING),
                            map(c => ({ creationOrError: c }))
                        )
                    )
                )
                .subscribe(stateUpdate => this.setState(stateUpdate))
        )
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        const disableForm = Boolean(
            this.state.creationOrError === LOADING ||
                (this.state.creationOrError && !isErrorLike(this.state.creationOrError)) ||
                isErrorLike(this.state.accountsOrError)
        )

        const selectedAccount: Account | undefined =
            this.state.accountsOrError !== LOADING && !isErrorLike(this.state.accountsOrError)
                ? this.state.accountsOrError.find(({ id }) => this.state.accountID === id)
                : undefined

        return (
            <div className="site-admin-create-product-subscription-page">
                <PageTitle title="Create product subscription" />
                <h2>Create product subscription</h2>
                <Form onSubmit={this.onSubmit}>
                    <div className="form-group">
                        <label htmlFor="site-admin-create-product-subscription-page__account">Account (customer)</label>
                        <select
                            id="site-admin-create-product-subscription-page__account"
                            className="form-control"
                            required={true}
                            disabled={
                                disableForm ||
                                this.state.accountsOrError === LOADING ||
                                isErrorLike(this.state.accountsOrError)
                            }
                            value={this.state.accountID}
                            onChange={this.onAccountIDChange}
                        >
                            {this.state.accountsOrError === LOADING ? (
                                <option value={this.state.accountID}>Loading...</option>
                            ) : (
                                !isErrorLike(this.state.accountsOrError) &&
                                this.state.accountsOrError.map(({ id, username, displayName }, i) => (
                                    <option key={i} value={id}>
                                        {username} {displayName && `(${displayName})`}
                                    </option>
                                ))
                            )}
                        </select>
                        {isErrorLike(this.state.accountsOrError) ? (
                            <ErrorAlert
                                className="mt-2"
                                error={this.state.accountsOrError}
                                prefix="Error loading accounts"
                            />
                        ) : selectedAccount ? (
                            <small className="form-text text-muted">
                                Email {pluralize('address', selectedAccount.emails.length, 'addresses')}:{' '}
                                <AccountEmailAddresses emails={selectedAccount.emails} />
                            </small>
                        ) : (
                            <small className="form-text text-muted">
                                The user associated with the customer account will be able to view this license key by
                                signing into Sourcegraph.com.
                            </small>
                        )}
                    </div>
                    <div className="form-group">
                        <button
                            type="submit"
                            disabled={
                                disableForm ||
                                this.state.accountsOrError === LOADING ||
                                isErrorLike(this.state.accountsOrError)
                            }
                            className={`btn btn-${disableForm ? 'secondary' : 'primary'}`}
                        >
                            Create product subscription
                        </button>
                        <small className="form-text text-muted">
                            You can generate a product license after creation.
                        </small>
                    </div>
                </Form>
                {isErrorLike(this.state.creationOrError) && (
                    <ErrorAlert className="mt-3" error={this.state.creationOrError} />
                )}
            </div>
        )
    }

    private onAccountIDChange: React.ChangeEventHandler<HTMLSelectElement> = e =>
        this.setState({ accountID: e.currentTarget.value })

    private onSubmit: React.FormEventHandler = e => {
        e.preventDefault()
        this.submits.next()
    }
}

function queryAccounts(): Observable<Account[]> {
    return queryGraphQL(
        gql`
            query ProductSubscriptionAccounts {
                users {
                    nodes {
                        id
                        username
                        displayName
                        emails {
                            email
                            verified
                        }
                    }
                }
            }
        `
    ).pipe(
        map(({ data, errors }) => {
            if (!data || !data.users || (errors && errors.length > 0)) {
                throw createAggregateError(errors)
            }
            return data.users.nodes
        })
    )
}

function createProductSubscription(
    args: GQL.ICreateProductSubscriptionOnDotcomMutationArguments
): Observable<Pick<GQL.IProductSubscription, 'id' | 'name' | 'url' | 'urlForSiteAdmin'>> {
    return mutateGraphQL(
        gql`
            mutation CreateProductSubscription($accountID: ID!) {
                dotcom {
                    createProductSubscription(accountID: $accountID) {
                        id
                        name
                        urlForSiteAdmin
                    }
                }
            }
        `,
        args
    ).pipe(
        map(({ data, errors }) => {
            if (!data || !data.dotcom || !data.dotcom.createProductSubscription || (errors && errors.length > 0)) {
                throw createAggregateError(errors)
            }
            return data.dotcom.createProductSubscription
        })
    )
}
