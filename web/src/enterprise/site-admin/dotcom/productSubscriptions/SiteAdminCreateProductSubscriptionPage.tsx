import * as H from 'history'
import AddIcon from 'mdi-react/AddIcon'
import React, { useCallback } from 'react'
import { Redirect, RouteComponentProps } from 'react-router'
import { Link } from 'react-router-dom'
import { concat, Observable } from 'rxjs'
import { catchError, concatMap, map, tap } from 'rxjs/operators'
import { dataOrThrowErrors, gql } from '../../../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { asError } from '../../../../../../shared/src/util/errors'
import { mutateGraphQL, queryGraphQL } from '../../../../backend/graphql'
import { FilteredConnection } from '../../../../components/FilteredConnection'
import { Form } from '../../../../components/Form'
import { PageTitle } from '../../../../components/PageTitle'
import { eventLogger } from '../../../../tracking/eventLogger'
import { useEventObservable } from '../../../../util/useObservable'

interface UserCreateSubscriptionNodeProps {
    /**
     * The user to display in this list item.
     */
    node: GQL.IUser

    /**
     * Browser history, used to redirect the user to the new subscription after one is successfully created.
     */
    history: H.History
}

const createProductSubscription = (
    args: GQL.ICreateProductSubscriptionOnDotcomMutationArguments
): Observable<Pick<GQL.IProductSubscription, 'urlForSiteAdmin'>> => {
    return mutateGraphQL(
        gql`
            mutation CreateProductSubscription($accountID: ID!) {
                dotcom {
                    createProductSubscription(accountID: $accountID) {
                        urlForSiteAdmin
                    }
                }
            }
        `,
        args
    ).pipe(
        map(dataOrThrowErrors),
        map(data => data.dotcom.createProductSubscription)
    )
}
const UserCreateSubscriptionNode: React.FunctionComponent<UserCreateSubscriptionNodeProps> = (
    props: UserCreateSubscriptionNodeProps
) => {
    const [onSubmit, createdSubscription] = useEventObservable(
        useCallback(
            (
                submits: Observable<React.FormEvent<HTMLFormElement>>
            ): Observable<Pick<GQL.IProductSubscription, 'urlForSiteAdmin'> | 'saving' | Error> =>
                submits.pipe(
                    tap(event => event.preventDefault()),
                    tap(() => eventLogger.log('NewProductSubscriptionCreated')),
                    map(() => ({ accountID: props.node.id })),
                    concatMap(input =>
                        concat(
                            ['saving' as const],
                            createProductSubscription(input).pipe(catchError(err => [asError(err)]))
                        )
                    )
                ),
            [props.node.id]
        )
    )

    return (
        <>
            {createdSubscription &&
                createdSubscription !== 'saving' &&
                !(createdSubscription instanceof Error) &&
                createdSubscription.urlForSiteAdmin && <Redirect to={createdSubscription.urlForSiteAdmin} />}
            <li className="list-group-item py-2">
                <div className="d-flex align-items-center justify-content-between">
                    <div>
                        <Link to={`/users/${props.node.username}`}>{props.node.username}</Link>{' '}
                        <span className="text-muted">
                            ({props.node.emails.filter(email => email.isPrimary).map(email => email.email)})
                        </span>
                    </div>
                    <div>
                        <Form onSubmit={onSubmit}>
                            <button
                                type="submit"
                                className="btn btn-sm btn-secondary"
                                disabled={createdSubscription === 'saving'}
                            >
                                <AddIcon className="icon-inline" /> Create new subscription
                            </button>
                        </Form>
                    </div>
                </div>
                {createdSubscription instanceof Error && (
                    <div className="alert alert-danger">{createdSubscription.message}</div>
                )}
                {createdSubscription &&
                    createdSubscription !== 'saving' &&
                    !(createdSubscription instanceof Error) &&
                    !createdSubscription.urlForSiteAdmin && (
                        <div className="alert alert-danger">
                            No subscription URL available (only accessible to site admins)
                        </div>
                    )}
            </li>
        </>
    )
}

class FilteredUserConnection extends FilteredConnection<GQL.IUser, Pick<UserCreateSubscriptionNodeProps, 'history'>> {}

interface Props extends RouteComponentProps<{}> {
    authenticatedUser: GQL.IUser
}

/**
 * Creates a product subscription for an account based on information provided in the displayed form.
 *
 * For use on Sourcegraph.com by Sourcegraph teammates only.
 */
export class SiteAdminCreateProductSubscriptionPage extends React.Component<Props> {
    public componentDidMount(): void {
        eventLogger.logViewEvent('SiteAdminCreateProductSubscription')
    }

    public render(): JSX.Element | null {
        const nodeProps: Pick<UserCreateSubscriptionNodeProps, 'history'> = {
            history: this.props.history,
        }
        return (
            <div className="site-admin-create-product-subscription-page">
                <PageTitle title="Create product subscription" />
                <h2>Create product subscription</h2>
                <FilteredUserConnection
                    className="list-group list-group-flush mt-3"
                    noun="user"
                    pluralNoun="users"
                    queryConnection={queryAccounts}
                    nodeComponent={UserCreateSubscriptionNode}
                    nodeComponentProps={nodeProps}
                    history={this.props.history}
                    location={this.props.location}
                />
            </div>
        )
    }
}

function queryAccounts(args: { first?: number; query?: string }): Observable<GQL.IUserConnection> {
    return queryGraphQL(
        gql`
            query ProductSubscriptionAccounts($first: Int, $query: String) {
                users(first: $first, query: $query) {
                    nodes {
                        id
                        username
                        emails {
                            email
                            verified
                            isPrimary
                        }
                    }
                }
            }
        `,
        args
    ).pipe(
        map(dataOrThrowErrors),
        map(data => data.users)
    )
}
