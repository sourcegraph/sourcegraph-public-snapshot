import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { gql, queryGraphQL } from '@sourcegraph/webapp/dist/backend/graphql'
import * as GQL from '@sourcegraph/webapp/dist/backend/graphqlschema'
import { CopyableText } from '@sourcegraph/webapp/dist/components/CopyableText'
import { PageTitle } from '@sourcegraph/webapp/dist/components/PageTitle'
import { SiteAdminAlert } from '@sourcegraph/webapp/dist/site-admin/SiteAdminAlert'
import { eventLogger } from '@sourcegraph/webapp/dist/tracking/eventLogger'
import { asError, createAggregateError, ErrorLike, isErrorLike } from '@sourcegraph/webapp/dist/util/errors'
import InformationIcon from 'mdi-react/InformationIcon'
import KeyIcon from 'mdi-react/KeyIcon'
import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { Link } from 'react-router-dom'
import { Observable, Subject, Subscription } from 'rxjs'
import { catchError, distinctUntilChanged, map, startWith, switchMap } from 'rxjs/operators'
import { ExpirationDate } from '../../productSubscription/ExpirationDate'
import { formatUserCount, mailtoSales } from '../../productSubscription/helpers'
import { LicenseGenerationKeyWarning } from '../../productSubscription/LicenseGenerationKeyWarning'
import { ProductCertificate } from '../../productSubscription/ProductCertificate'
import { BackToAllSubscriptionsLink } from './BackToAllSubscriptionsLink'
import { ProductSubscriptionBilling } from './ProductSubscriptionBilling'
import { ProductSubscriptionHistory } from './ProductSubscriptionHistory'

interface Props extends RouteComponentProps<{ subscriptionID: string }> {
    user: GQL.IUser
}

const LOADING: 'loading' = 'loading'

interface State {
    showLicenseKey: boolean

    /**
     * The product subscription, or loading, or an error.
     */
    productSubscriptionOrError: typeof LOADING | GQL.IProductSubscription | ErrorLike
}

/**
 * Displays a product subscription in the user subscriptions area.
 */
export class UserSubscriptionsProductSubscriptionPage extends React.Component<Props, State> {
    public state: State = {
        showLicenseKey: false,
        productSubscriptionOrError: LOADING,
    }

    private componentUpdates = new Subject<Props>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        eventLogger.logViewEvent('UserSubscriptionsProductSubscription')

        const subscriptionIDChanges = this.componentUpdates.pipe(
            map(props => props.match.params.subscriptionID),
            distinctUntilChanged()
        )

        const productSubscriptionChanges = subscriptionIDChanges.pipe(
            switchMap(subscriptionID =>
                this.queryProductSubscription(subscriptionID).pipe(
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

        this.componentUpdates.next(this.props)
    }

    public componentDidUpdate(): void {
        this.componentUpdates.next(this.props)
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        return (
            <div className="user-subscriptions-product-subscription-page">
                <PageTitle title="Subscription" />
                <div className="d-flex align-items-center justify-content-between">
                    <BackToAllSubscriptionsLink user={this.props.user} />
                    {this.state.productSubscriptionOrError !== LOADING &&
                        !isErrorLike(this.state.productSubscriptionOrError) &&
                        this.state.productSubscriptionOrError.urlForSiteAdmin && (
                            <SiteAdminAlert className="small m-0 p-1">
                                <Link
                                    to={this.state.productSubscriptionOrError.urlForSiteAdmin}
                                    className="mt-2 d-block"
                                >
                                    View subscription
                                </Link>
                            </SiteAdminAlert>
                        )}
                </div>
                {this.state.productSubscriptionOrError === LOADING ? (
                    <LoadingSpinner className="icon-inline" />
                ) : isErrorLike(this.state.productSubscriptionOrError) ? (
                    <div className="alert alert-danger my-2">
                        Error: {this.state.productSubscriptionOrError.message}
                    </div>
                ) : (
                    <div className="row">
                        <div className="col-md-9">
                            <h2>Subscription {this.state.productSubscriptionOrError.name}</h2>
                            {this.state.productSubscriptionOrError.plan &&
                                this.state.productSubscriptionOrError.userCount &&
                                this.state.productSubscriptionOrError.expiresAt && (
                                    <ProductCertificate
                                        title={this.state.productSubscriptionOrError.plan.fullProductName}
                                        subtitle={
                                            <>
                                                {formatUserCount(this.state.productSubscriptionOrError.userCount, true)}{' '}
                                                license,{' '}
                                                <ExpirationDate
                                                    date={this.state.productSubscriptionOrError.expiresAt}
                                                    showRelative={true}
                                                    lowercase={true}
                                                />
                                            </>
                                        }
                                        footer={
                                            <>
                                                <div className="card-footer d-flex align-items-center justify-content-between flex-wrap">
                                                    <button
                                                        type="button"
                                                        className="btn btn-primary mr-4 my-1"
                                                        onClick={this.toggleShowLicenseKey}
                                                    >
                                                        <KeyIcon className="icon-inline" />{' '}
                                                        {this.state.showLicenseKey ? 'Hide' : 'Reveal'} license key
                                                    </button>
                                                    <div className="flex-fill" />
                                                    <div className="my-1" />
                                                </div>
                                                {this.state.showLicenseKey && (
                                                    <div className="card-footer">
                                                        <h3>License key</h3>
                                                        {this.state.productSubscriptionOrError.activeLicense ? (
                                                            <>
                                                                <CopyableText
                                                                    text={
                                                                        this.state.productSubscriptionOrError
                                                                            .activeLicense.licenseKey
                                                                    }
                                                                    className="d-block"
                                                                />
                                                                <small className="mt-2 d-flex align-items-center">
                                                                    <InformationIcon className="icon-inline mr-1" />{' '}
                                                                    <span>
                                                                        Use this license key as the{' '}
                                                                        <code>
                                                                            <strong>licenseKey</strong>
                                                                        </code>{' '}
                                                                        property value in Sourcegraph site
                                                                        configuration.
                                                                    </span>
                                                                </small>
                                                                <LicenseGenerationKeyWarning className="mb-0 mt-1" />
                                                            </>
                                                        ) : (
                                                            <div className="text-muted">
                                                                No license key found.{' '}
                                                                <a
                                                                    href={mailtoSales({
                                                                        subject: `No license key for subscription ${
                                                                            this.state.productSubscriptionOrError.name
                                                                        }`,
                                                                    })}
                                                                >
                                                                    Contact sales
                                                                </a>{' '}
                                                                for help.
                                                            </div>
                                                        )}
                                                    </div>
                                                )}
                                            </>
                                        }
                                    />
                                )}
                            <div className="card mt-3">
                                <div className="card-header">Billing</div>
                                <ProductSubscriptionBilling
                                    productSubscription={this.state.productSubscriptionOrError}
                                />
                                <div className="card-footer">
                                    <a
                                        href={mailtoSales({
                                            subject: `No license key for subscription ${
                                                this.state.productSubscriptionOrError.name
                                            }`,
                                        })}
                                    >
                                        Contact sales
                                    </a>{' '}
                                    to change your payment method.
                                </div>
                            </div>
                            <div className="card mt-3">
                                <div className="card-header">History</div>
                                <ProductSubscriptionHistory
                                    productSubscription={this.state.productSubscriptionOrError}
                                />
                            </div>
                        </div>
                    </div>
                )}
            </div>
        )
    }

    private toggleShowLicenseKey = () => this.setState(prevState => ({ showLicenseKey: !prevState.showLicenseKey }))

    private queryProductSubscription = (id: GQL.ID): Observable<GQL.IProductSubscription> =>
        queryGraphQL(
            gql`
                query ProductSubscription($id: ID!) {
                    node(id: $id) {
                        ... on ProductSubscription {
                            ...ProductSubscriptionFields
                        }
                    }
                }
                fragment ProductSubscriptionFields on ProductSubscription {
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
                    plan {
                        billingID
                        name
                        title
                        fullProductName
                        pricePerUserPerYear
                    }
                    userCount
                    expiresAt
                    events {
                        id
                        date
                        title
                        description
                        url
                    }
                    activeLicense {
                        licenseKey
                        info {
                            fullProductName
                            tags
                            userCount
                            expiresAt
                        }
                    }
                    createdAt
                    isArchived
                    url
                    urlForSiteAdmin
                }
            `,
            { id }
        ).pipe(
            map(({ data, errors }) => {
                if (!data || !data.node || (errors && errors.length > 0)) {
                    throw createAggregateError(errors)
                }
                return data.node as GQL.IProductSubscription
            })
        )
}
