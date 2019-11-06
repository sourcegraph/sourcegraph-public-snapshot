import addDays from 'date-fns/addDays'
import endOfDay from 'date-fns/endOfDay'
import { upperFirst } from 'lodash'
import * as React from 'react'
import { Observable, Subject, Subscription } from 'rxjs'
import { catchError, map, startWith, switchMap, tap } from 'rxjs/operators'
import { gql } from '../../../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { asError, createAggregateError, ErrorLike, isErrorLike } from '../../../../../../shared/src/util/errors'
import { mutateGraphQL } from '../../../../backend/graphql'
import { Form } from '../../../../components/Form'
import { ExpirationDate } from '../../../productSubscription/ExpirationDate'

interface Props {
    subscriptionID: GQL.ID
    onGenerate: () => void
}

const LOADING: 'loading' = 'loading'

interface State {
    /** Comma-separated license tags. */
    tags: string

    userCount: number
    validDays: number | null
    expiresAt: number | null

    /**
     * The result of creating the product subscription, or null when not pending or complete, or loading, or an
     * error.
     */
    creationOrError: null | Pick<GQL.IProductSubscription, 'id'> | typeof LOADING | ErrorLike
}

/**
 * Displays a form to generate a new product license for a product subscription.
 *
 * For use on Sourcegraph.com by Sourcegraph teammates only.
 */
export class SiteAdminGenerateProductLicenseForSubscriptionForm extends React.Component<Props, State> {
    private get emptyState(): Pick<State, 'tags' | 'userCount' | 'validDays' | 'expiresAt' | 'creationOrError'> {
        return {
            tags: '',
            userCount: 1,
            validDays: 1,
            expiresAt: addDaysAndRoundToEndOfDay(1),
            creationOrError: null,
        }
    }

    private static DURATION_LINKS = [
        { label: '7 days', days: 7 },
        { label: '14 days', days: 14 },
        { label: '30 days', days: 30 },
        { label: '60 days', days: 60 },
        { label: '1 year', days: 366 }, // 366 not 365 to account for leap year
    ]

    public state: State = {
        ...this.emptyState,
        tags: 'true-up', // Default because we expect most licenses will be true-up.
    }

    private submits = new Subject<void>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        this.subscriptions.add(
            this.submits
                .pipe(
                    switchMap(() => {
                        if (this.state.expiresAt === null) {
                            throw new Error('invalid expiresAt')
                        }
                        return generateProductLicenseForSubscription({
                            productSubscriptionID: this.props.subscriptionID,
                            license: {
                                tags: this.state.tags ? this.state.tags.split(',') : [],
                                userCount: this.state.userCount,
                                expiresAt: Math.ceil(this.state.expiresAt / 1000),
                            },
                        }).pipe(
                            tap(() => this.props.onGenerate()),
                            catchError(err => [asError(err)]),
                            startWith(LOADING),
                            map(c => ({ creationOrError: c }))
                        )
                    })
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
                (this.state.creationOrError && !isErrorLike(this.state.creationOrError))
        )

        return (
            <div className="site-admin-generate-product-license-for-subscription-form">
                {this.state.creationOrError &&
                !isErrorLike(this.state.creationOrError) &&
                this.state.creationOrError !== LOADING ? (
                    <div className="border rounded border-success mb-5">
                        <div className="border-top-0 border-left-0 border-right-0 rounded-0 alert alert-success mb-0 d-flex align-items-center justify-content-between px-3 py-2">
                            <span>Generated product license.</span>
                            <button
                                type="button"
                                className="btn btn-primary"
                                onClick={this.dismissAlert}
                                autoFocus={true}
                            >
                                Dismiss
                            </button>
                        </div>
                    </div>
                ) : (
                    <Form onSubmit={this.onSubmit}>
                        <div className="form-group">
                            <label htmlFor="site-admin-create-product-subscription-page__tags">Tags</label>
                            <input
                                type="text"
                                className="form-control"
                                id="site-admin-create-product-subscription-page__tags"
                                disabled={disableForm}
                                value={this.state.tags}
                                list="knownPlans"
                                onChange={this.onPlanChange}
                            />
                            <datalist id="knownPlans">
                                <option value="true-up" />
                                <option value="trial" />
                                <option value="starter,trial" />
                                <option value="starter,true-up" />
                                <option value="dev" />
                            </datalist>
                            <small className="form-text text-muted">
                                Tags restrict a license. Recognized tags: <code>true-up</code> (allow user creation
                                beyond limit), <code>starter</code> (Enterprise Starter), <code>trial</code> (shows a
                                "trial" indicator), <code>dev</code> (shows a "for development only" indicator).
                                Separate multiple with commas and no spaces (like <code>starter,trial</code>
                                ). Order does not matter.
                            </small>
                            <small className="form-text text-muted mt-2">
                                To find the exact license tags used for licenses generated by self-service payment, view
                                the <code>licenseTags</code> product plan metadata item in the billing system.
                            </small>
                        </div>
                        <div className="form-group">
                            <label htmlFor="site-admin-create-product-subscription-page__userCount">Users</label>
                            <input
                                type="number"
                                min={1}
                                className="form-control"
                                id="site-admin-create-product-subscription-page__userCount"
                                disabled={disableForm}
                                value={this.state.userCount || ''}
                                onChange={this.onUserCountChange}
                            />
                        </div>
                        <div className="form-group">
                            <label htmlFor="site-admin-create-product-subscription-page__validDays">
                                Valid for (days)
                            </label>
                            <input
                                type="number"
                                className="form-control"
                                id="site-admin-create-product-subscription-page__validDays"
                                disabled={disableForm}
                                value={this.state.validDays || ''}
                                min={1}
                                max={2000} // avoid overflowing int32
                                onChange={this.onValidDaysChange}
                            />
                            <small className="form-text text-muted">
                                {this.state.expiresAt !== null ? (
                                    <ExpirationDate
                                        date={this.state.expiresAt}
                                        showTime={true}
                                        showRelative={true}
                                        showPrefix={true}
                                    />
                                ) : (
                                    <>&nbsp;</>
                                )}
                            </small>
                            <small className="form-text text-muted d-block mt-1">
                                Set to{' '}
                                {SiteAdminGenerateProductLicenseForSubscriptionForm.DURATION_LINKS.map(
                                    ({ label, days }) => (
                                        <a
                                            href="#"
                                            key={days}
                                            className="mr-2"
                                            onClick={e => {
                                                e.preventDefault()
                                                this.setValidDays(days)
                                            }}
                                        >
                                            {label}
                                        </a>
                                    )
                                )}
                            </small>
                        </div>
                        <button
                            type="submit"
                            disabled={disableForm}
                            className={`btn btn-${disableForm ? 'secondary' : 'primary'}`}
                        >
                            Generate license
                        </button>
                    </Form>
                )}
                {isErrorLike(this.state.creationOrError) && (
                    <div className="alert alert-danger mt-3">{upperFirst(this.state.creationOrError.message)}</div>
                )}
            </div>
        )
    }

    private onPlanChange: React.ChangeEventHandler<HTMLInputElement> = e =>
        this.setState({ tags: e.currentTarget.value })

    private onUserCountChange: React.ChangeEventHandler<HTMLInputElement> = e =>
        this.setState({ userCount: e.currentTarget.valueAsNumber })

    private onValidDaysChange: React.ChangeEventHandler<HTMLInputElement> = e =>
        this.setValidDays(Number.isNaN(e.currentTarget.valueAsNumber) ? null : e.currentTarget.valueAsNumber)

    private onSubmit: React.FormEventHandler = e => {
        e.preventDefault()
        this.submits.next()
    }

    private setValidDays(validDays: number | null): void {
        this.setState({
            validDays,
            expiresAt: validDays !== null ? addDaysAndRoundToEndOfDay(validDays || 0) : null,
        })
    }

    private dismissAlert = (): void => this.setState(this.emptyState)
}

/**
 * Adds 1 day to the current date, then rounds it up to midnight in the client's timezone. This is a
 * generous interpretation of "valid for N days" to avoid confusion over timezones or "will it
 * expire at the beginning of the day or at the end of the day?"
 */
function addDaysAndRoundToEndOfDay(amount: number): number {
    return endOfDay(addDays(Date.now(), amount)).getTime()
}

function generateProductLicenseForSubscription(
    args: GQL.IGenerateProductLicenseForSubscriptionOnDotcomMutationArguments
): Observable<Pick<GQL.IProductSubscription, 'id'>> {
    return mutateGraphQL(
        gql`
            mutation GenerateProductLicenseForSubscription(
                $productSubscriptionID: ID!
                $license: ProductLicenseInput!
            ) {
                dotcom {
                    generateProductLicenseForSubscription(
                        productSubscriptionID: $productSubscriptionID
                        license: $license
                    ) {
                        id
                    }
                }
            }
        `,
        args
    ).pipe(
        map(({ data, errors }) => {
            if (
                !data ||
                !data.dotcom ||
                !data.dotcom.generateProductLicenseForSubscription ||
                (errors && errors.length > 0)
            ) {
                throw createAggregateError(errors)
            }
            return data.dotcom.generateProductLicenseForSubscription
        })
    )
}
