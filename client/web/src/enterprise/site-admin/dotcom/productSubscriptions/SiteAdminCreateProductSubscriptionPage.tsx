import React, { useCallback, useEffect } from 'react'

import { mdiPlus } from '@mdi/js'
import { Navigate } from 'react-router-dom'
import { merge, of, type Observable } from 'rxjs'
import { catchError, concatMapTo, map, tap } from 'rxjs/operators'

import { asError, type ErrorLike, isErrorLike } from '@sourcegraph/common'
import { dataOrThrowErrors, gql } from '@sourcegraph/http-client'
import { Button, useEventObservable, Link, Alert, Icon, H2, Form } from '@sourcegraph/wildcard'

import type { AuthenticatedUser } from '../../../../auth'
import { mutateGraphQL, queryGraphQL } from '../../../../backend/graphql'
import { FilteredConnection } from '../../../../components/FilteredConnection'
import { PageTitle } from '../../../../components/PageTitle'
import type {
    CreateProductSubscriptionVariables,
    ProductSubscriptionAccountsResult,
    ProductSubscriptionAccountsVariables,
    ProductSubscriptionAccountFields,
    CreateProductSubscriptionResult,
} from '../../../../graphql-operations'
import { eventLogger } from '../../../../tracking/eventLogger'

interface UserCreateSubscriptionNodeProps {
    /**
     * The user to display in this list item.
     */
    node: ProductSubscriptionAccountFields
}

const createProductSubscription = (
    args: CreateProductSubscriptionVariables
): Observable<CreateProductSubscriptionResult['dotcom']['createProductSubscription']> =>
    mutateGraphQL<CreateProductSubscriptionResult>(
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
            ): Observable<
                CreateProductSubscriptionResult['dotcom']['createProductSubscription'] | 'saving' | ErrorLike
            > =>
                submits.pipe(
                    tap(event => event.preventDefault()),
                    tap(() => {
                        window.context.telemetryRecorder?.recordEvent('newProductSubscription', 'created')
                        eventLogger.log('NewProductSubscriptionCreated')
                    }),
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
                createdSubscription.urlForSiteAdmin && (
                    <Navigate replace={true} to={createdSubscription.urlForSiteAdmin} />
                )}
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
                                <Icon aria-hidden={true} svgPath={mdiPlus} /> Create new subscription
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

interface Props {
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
        window.context.telemetryRecorder?.recordEvent('siteAdminCreateProductSubscription', 'viewed')
        eventLogger.logViewEvent('SiteAdminCreateProductSubscription')
    })
    return (
        <div className="site-admin-create-product-subscription-page">
            <PageTitle title="Create product subscription" />
            <H2>Create product subscription</H2>
            <FilteredConnection<ProductSubscriptionAccountFields, {}>
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

function queryAccounts(
    args: Partial<ProductSubscriptionAccountsVariables>
): Observable<ProductSubscriptionAccountsResult['users']> {
    return queryGraphQL<ProductSubscriptionAccountsResult>(
        gql`
            query ProductSubscriptionAccounts($first: Int, $query: String) {
                users(first: $first, query: $query) {
                    nodes {
                        ...ProductSubscriptionAccountFields
                    }
                    totalCount
                    pageInfo {
                        hasNextPage
                    }
                }
            }
            fragment ProductSubscriptionAccountFields on User {
                id
                username
                emails {
                    email
                    verified
                    isPrimary
                }
            }
        `,
        args
    ).pipe(
        map(dataOrThrowErrors),
        map(data => data.users)
    )
}
