import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { upperFirst } from 'lodash'
import AddIcon from 'mdi-react/AddIcon'
import ArrowLeftIcon from 'mdi-react/ArrowLeftIcon'
import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { Link } from 'react-router-dom'
import { combineLatest, Observable, Subject, Subscription } from 'rxjs'
import {
    catchError,
    distinctUntilChanged,
    filter,
    map,
    mapTo,
    startWith,
    switchMap,
    tap,
    withLatestFrom,
} from 'rxjs/operators'
import { gql } from '../../../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { asError, createAggregateError, ErrorLike, isErrorLike } from '../../../../../../shared/src/util/errors'
import { mutateGraphQL, queryGraphQL } from '../../../../backend/graphql'
import { FilteredConnection } from '../../../../components/FilteredConnection'
import { PageTitle } from '../../../../components/PageTitle'
import { Timestamp } from '../../../../components/time/Timestamp'
import { eventLogger } from '../../../../tracking/eventLogger'
import { AccountEmailAddresses } from '../../../dotcom/productSubscriptions/AccountEmailAddresses'
import { AccountName } from '../../../dotcom/productSubscriptions/AccountName'
import { ProductSubscriptionLabel } from '../../../dotcom/productSubscriptions/ProductSubscriptionLabel'
import { LicenseGenerationKeyWarning } from '../../../productSubscription/LicenseGenerationKeyWarning'
import { ProductSubscriptionHistory } from '../../../user/productSubscriptions/ProductSubscriptionHistory'
import { SiteAdminGenerateProductLicenseForSubscriptionForm } from './SiteAdminGenerateProductLicenseForSubscriptionForm'
import {
    siteAdminProductLicenseFragment,
    SiteAdminProductLicenseNode,
    SiteAdminProductLicenseNodeProps,
} from './SiteAdminProductLicenseNode'
import { SiteAdminProductSubscriptionBillingLink } from './SiteAdminProductSubscriptionBillingLink'

interface Props extends RouteComponentProps<{ subscriptionUUID: string }> {}

class FilteredSiteAdminProductLicenseConnection extends FilteredConnection<
    GQL.IProductLicense,
    Pick<SiteAdminProductLicenseNodeProps, 'onDidUpdate' | 'showSubscription'>
> {}

const LOADING: 'loading' = 'loading'

interface State {
    showGenerate: boolean

    /**
     * The product subscription, or loading, or an error.
     */
    productSubscriptionOrError: typeof LOADING | GQL.IProductSubscription | ErrorLike

    /** The result of archiving this subscription: null for done or not started, loading, or an error. */
    archivalOrError: typeof LOADING | null | ErrorLike
}

/**
 * Displays a product subscription in the site admin area.
 */
export class SiteAdminProductSubscriptionPage extends React.Component<Props, State> {
    public state: State = {
        showGenerate: false,
        productSubscriptionOrError: LOADING,
        archivalOrError: null,
    }

    private componentUpdates = new Subject<Props>()
    private archivals = new Subject<void>()
    private licenseUpdates = new Subject<void>()
    private updates = new Subject<void>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        eventLogger.logViewEvent('SiteAdminProductSubscription')

        const subscriptionUUIDChanges = this.componentUpdates.pipe(
            map(props => props.match.params.subscriptionUUID),
            distinctUntilChanged()
        )

        const productSubscriptionChanges = combineLatest([
            subscriptionUUIDChanges,
            this.updates.pipe(startWith(undefined)),
        ]).pipe(
            switchMap(([subscriptionUUID]) =>
                this.queryProductSubscription(subscriptionUUID).pipe(
                    catchError(err => [asError(err)]),
                    startWith(LOADING)
                )
            )
        )

        this.subscriptions.add(
            productSubscriptionChanges
                .pipe(map(result => ({ productSubscriptionOrError: result })))
                .subscribe(stateUpdate => this.setState(stateUpdate))
        )

        this.subscriptions.add(
            this.archivals
                .pipe(
                    withLatestFrom(
                        productSubscriptionChanges.pipe(
                            filter((v): v is GQL.IProductSubscription => v !== LOADING && !isErrorLike(v))
                        )
                    ),
                    filter(() =>
                        window.confirm(
                            'Really archive this product subscription? This will hide it from site admins and users.\n\nHowever, it does NOT:\n\n- invalidate the license key\n- refund payment or cancel billing\n\nYou must manually do those things.'
                        )
                    ),
                    switchMap(([, { id }]) =>
                        archiveProductSubscription({ id }).pipe(
                            mapTo(null),
                            tap(() => this.props.history.push('/site-admin/dotcom/product/subscriptions')),
                            catchError(error => [asError(error)]),
                            map(c => ({ archivalOrError: c })),
                            startWith({ archivalOrError: LOADING })
                        )
                    )
                )
                .subscribe(stateUpdate => this.setState(stateUpdate), error => console.error(error))
        )

        this.componentUpdates.next(this.props)
    }

    public componentDidUpdate(): void {
        this.componentUpdates.next(this.props)
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        const nodeProps: Pick<SiteAdminProductLicenseNodeProps, 'onDidUpdate' | 'showSubscription'> = {
            onDidUpdate: this.onDidUpdateProductLicense,
            showSubscription: false,
        }

        return (
            <div className="site-admin-product-subscription-page">
                <PageTitle title="Product subscription" />
                <div className="mb-2">
                    <Link to="/site-admin/dotcom/product/subscriptions" className="btn btn-link btn-sm">
                        <ArrowLeftIcon className="icon-inline" /> All subscriptions
                    </Link>
                </div>
                {this.state.productSubscriptionOrError === LOADING ? (
                    <LoadingSpinner className="icon-inline" />
                ) : isErrorLike(this.state.productSubscriptionOrError) ? (
                    <div className="alert alert-danger my-2">
                        Error: {this.state.productSubscriptionOrError.message}
                    </div>
                ) : (
                    <>
                        <h2>Product subscription {this.state.productSubscriptionOrError.name}</h2>
                        <div className="mb-3">
                            <button
                                type="button"
                                className="btn btn-danger"
                                onClick={this.archiveProductSubscription}
                                disabled={this.state.archivalOrError === null}
                            >
                                Archive
                            </button>
                            {isErrorLike(this.state.archivalOrError) && (
                                <div className="alert alert-danger mt-2">
                                    Error: {upperFirst(this.state.archivalOrError.message)}
                                </div>
                            )}
                        </div>
                        <div className="card mt-3">
                            <div className="card-header">Details</div>
                            <table className="table mb-0">
                                <tbody>
                                    <tr>
                                        <th className="text-nowrap">ID</th>
                                        <td className="w-100">{this.state.productSubscriptionOrError.name}</td>
                                    </tr>
                                    <tr>
                                        <th className="text-nowrap">Plan</th>
                                        <td className="w-100">
                                            <ProductSubscriptionLabel
                                                productSubscription={this.state.productSubscriptionOrError}
                                            />
                                        </td>
                                    </tr>
                                    <tr>
                                        <th className="text-nowrap">Account</th>
                                        <td className="w-100">
                                            <AccountName account={this.state.productSubscriptionOrError.account} />{' '}
                                            &mdash;{' '}
                                            <Link to={this.state.productSubscriptionOrError.url}>View as user</Link>
                                        </td>
                                    </tr>
                                    <tr>
                                        <th className="text-nowrap">Account emails</th>
                                        <td className="w-100">
                                            {this.state.productSubscriptionOrError.account && (
                                                <AccountEmailAddresses
                                                    emails={this.state.productSubscriptionOrError.account.emails}
                                                />
                                            )}
                                        </td>
                                    </tr>
                                    <tr>
                                        <th className="text-nowrap">Billing</th>
                                        <td className="w-100">
                                            <SiteAdminProductSubscriptionBillingLink
                                                productSubscription={this.state.productSubscriptionOrError}
                                                onDidUpdate={this.onDidUpdate}
                                            />
                                        </td>
                                    </tr>
                                    <tr>
                                        <th className="text-nowrap">Created at</th>
                                        <td className="w-100">
                                            <Timestamp date={this.state.productSubscriptionOrError.createdAt} />
                                        </td>
                                    </tr>
                                </tbody>
                            </table>
                        </div>
                        <LicenseGenerationKeyWarning className="mt-3" />
                        <div className="card mt-1">
                            <div className="card-header d-flex align-items-center justify-content-between">
                                Licenses
                                {this.state.showGenerate ? (
                                    <button
                                        type="button"
                                        className="btn btn-secondary"
                                        onClick={this.toggleShowGenerate}
                                    >
                                        Dismiss new license form
                                    </button>
                                ) : (
                                    <button
                                        type="button"
                                        className="btn btn-primary btn-sm"
                                        onClick={this.toggleShowGenerate}
                                    >
                                        <AddIcon className="icon-inline" /> Generate new license manually
                                    </button>
                                )}
                            </div>
                            {this.state.showGenerate && (
                                <div className="card-body">
                                    <SiteAdminGenerateProductLicenseForSubscriptionForm
                                        subscriptionID={this.state.productSubscriptionOrError.id}
                                        onGenerate={this.onDidUpdateProductLicense}
                                    />
                                </div>
                            )}
                            <FilteredSiteAdminProductLicenseConnection
                                className="list-group list-group-flush"
                                noun="product license"
                                pluralNoun="product licenses"
                                queryConnection={this.queryProductLicenses}
                                nodeComponent={SiteAdminProductLicenseNode}
                                nodeComponentProps={nodeProps}
                                compact={true}
                                hideSearch={true}
                                noSummaryIfAllNodesVisible={true}
                                updates={this.licenseUpdates}
                                history={this.props.history}
                                location={this.props.location}
                            />
                        </div>
                        <div className="card mt-3">
                            <div className="card-header">History</div>
                            <ProductSubscriptionHistory productSubscription={this.state.productSubscriptionOrError} />
                        </div>
                    </>
                )}
            </div>
        )
    }

    private toggleShowGenerate = (): void => this.setState(prevState => ({ showGenerate: !prevState.showGenerate }))

    private queryProductSubscription = (uuid: string): Observable<GQL.IProductSubscription> =>
        queryGraphQL(
            gql`
                query ProductSubscription($uuid: String!) {
                    dotcom {
                        productSubscription(uuid: $uuid) {
                            id
                            name
                            account {
                                id
                                username
                                displayName
                                emails {
                                    email
                                    verified
                                }
                            }
                            invoiceItem {
                                plan {
                                    billingPlanID
                                    name
                                    nameWithBrand
                                    pricePerUserPerYear
                                }
                                userCount
                                expiresAt
                            }
                            events {
                                id
                                date
                                title
                                description
                                url
                            }
                            productLicenses {
                                nodes {
                                    id
                                    info {
                                        tags
                                        userCount
                                        expiresAt
                                    }
                                    licenseKey
                                    createdAt
                                }
                                totalCount
                                pageInfo {
                                    hasNextPage
                                }
                            }
                            createdAt
                            isArchived
                            url
                            urlForSiteAdminBilling
                        }
                    }
                }
            `,
            { uuid }
        ).pipe(
            map(({ data, errors }) => {
                if (!data || !data.dotcom || !data.dotcom.productSubscription || (errors && errors.length > 0)) {
                    throw createAggregateError(errors)
                }
                return data.dotcom.productSubscription
            })
        )

    private queryProductLicenses = (args: { first?: number }): Observable<GQL.IProductLicenseConnection> =>
        queryGraphQL(
            gql`
                query ProductLicenses($first: Int, $subscriptionUUID: String!) {
                    dotcom {
                        productSubscription(uuid: $subscriptionUUID) {
                            productLicenses(first: $first) {
                                nodes {
                                    ...ProductLicenseFields
                                }
                                totalCount
                                pageInfo {
                                    hasNextPage
                                }
                            }
                        }
                    }
                }
                ${siteAdminProductLicenseFragment}
            `,
            {
                first: args.first,
                subscriptionUUID: this.props.match.params.subscriptionUUID,
            }
        ).pipe(
            map(({ data, errors }) => {
                if (
                    !data ||
                    !data.dotcom ||
                    !data.dotcom.productSubscription ||
                    !data.dotcom.productSubscription.productLicenses ||
                    (errors && errors.length > 0)
                ) {
                    throw createAggregateError(errors)
                }
                return data.dotcom.productSubscription.productLicenses
            })
        )

    private archiveProductSubscription = (): void => this.archivals.next()

    private onDidUpdateProductLicense = (): void => {
        this.licenseUpdates.next()
        this.toggleShowGenerate()
    }

    private onDidUpdate = (): void => this.updates.next()
}

function archiveProductSubscription(args: GQL.IArchiveProductSubscriptionOnDotcomMutationArguments): Observable<void> {
    return mutateGraphQL(
        gql`
            mutation ArchiveProductSubscription($id: ID!) {
                dotcom {
                    archiveProductSubscription(id: $id) {
                        alwaysNil
                    }
                }
            }
        `,
        args
    ).pipe(
        map(({ data, errors }) => {
            if (!data || !data.dotcom || !data.dotcom.archiveProductSubscription || (errors && errors.length > 0)) {
                throw createAggregateError(errors)
            }
        })
    )
}
