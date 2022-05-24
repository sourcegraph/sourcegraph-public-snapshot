import React, { useCallback, useEffect } from 'react'

import * as H from 'history'
import AddIcon from 'mdi-react/AddIcon'
import { Redirect, RouteComponentProps } from 'react-router'
import { merge, of, Observable } from 'rxjs'
import { catchError, concatMapTo, map, tap } from 'rxjs/operators'

import { Form } from '@sourcegraph/branded/src/components/Form'
import { asError, ErrorLike, isErrorLike } from '@sourcegraph/common'
import { dataOrThrowErrors, gql } from '@sourcegraph/http-client'
import * as GQL from '@sourcegraph/shared/src/schema'
import { Button, useEventObservable, Link, Alert, Icon, Typography } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../../../../auth'
import { mutateGraphQL, queryGraphQL } from '../../../../backend/graphql'
import { FilteredConnection } from '../../../../components/FilteredConnection'
import { PageTitle } from '../../../../components/PageTitle'
import { eventLogger } from '../../../../tracking/eventLogger'

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
): Observable<Pick<GQL.IProductSubscription, 'urlForSiteAdmin'>> =>
    mutateGraphQL(
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

const UserCreateSubscriptionNode: React.FunctionComponent<React.PropsWithChildren<UserCreateSubscriptionNodeProps>> = (
    props: UserCreateSubscriptionNodeProps
) => {
    const [onSubmit, createdSubscription] = useEventObservable(
        useCallback(
            (
                submits: Observable<React.FormEvent<HTMLFormElement>>
            ): Observable<Pick<GQL.IProductSubscription, 'urlForSiteAdmin'> | 'saving' | ErrorLike> =>
                submits.pipe(
                    tap(event => event.preventDefault()),
                    tap(() => eventLogger.log('NewProductSubscriptionCreated')),
                    concatMapTo(
                        merge(
                            of('saving' as const),
                            createProductSubscription({ accountID: props.node.id }).pipe(
                                catchError(error => [asError(error)])
                            )
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
                !isErrorLike(createdSubscription) &&
                createdSubscription.urlForSiteAdmin && <Redirect to={createdSubscription.urlForSiteAdmin} />}
            <li className="list-group-item py-2">
                <div className="d-flex align-items-center justify-content-between">
                    <div>
                        <Link to={`/users/${props.node.username}`}>{props.node.username}</Link>{' '}
                        <span className="text-muted">
                            ({props.node.emails.filter(({ isPrimary }) => isPrimary).map(({ email }) => email)})
                        </span>
                    </div>
                    <div>
                        <Form onSubmit={onSubmit}>
                            <Button
                                type="submit"
                                disabled={createdSubscription === 'saving'}
                                variant="secondary"
                                size="sm"
                            >
                                <Icon as={AddIcon} /> Create new subscription
                            </Button>
                        </Form>
                    </div>
                </div>
                {isErrorLike(createdSubscription) && <Alert variant="danger">{createdSubscription.message}</Alert>}
                {createdSubscription &&
                    createdSubscription !== 'saving' &&
                    !isErrorLike(createdSubscription) &&
                    !createdSubscription.urlForSiteAdmin && (
                        <Alert variant="danger">No subscription URL available (only accessible to site admins)</Alert>
                    )}
            </li>
        </>
    )
}

class FilteredUserConnection extends FilteredConnection<GQL.IUser, Pick<UserCreateSubscriptionNodeProps, 'history'>> {}

interface Props extends RouteComponentProps<{}> {
    authenticatedUser: AuthenticatedUser
}

/**
 * Creates a product subscription for an account based on information provided in the displayed form.
 *
 * For use on Sourcegraph.com by Sourcegraph teammates only.
 */
export const SiteAdminCreateProductSubscriptionPage: React.FunctionComponent<
    React.PropsWithChildren<Props>
> = props => {
    useEffect(() => {
        eventLogger.logViewEvent('SiteAdminCreateProductSubscription')
    })
    return (
        <div className="site-admin-create-product-subscription-page">
            <PageTitle title="Create product subscription" />
            <Typography.H2>Create product subscription</Typography.H2>
            <FilteredUserConnection
                {...props}
                className="list-group list-group-flush mt-3"
                noun="user"
                pluralNoun="users"
                queryConnection={queryAccounts}
                nodeComponent={UserCreateSubscriptionNode}
                nodeComponentProps={props}
            />
        </div>
    )
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
                    totalCount
                    pageInfo {
                        hasNextPage
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
