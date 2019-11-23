import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import * as React from 'react'
import { Observable, Subscription } from 'rxjs'
import { catchError, map, startWith, tap } from 'rxjs/operators'
import { gql } from '../../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { asError, createAggregateError, ErrorLike, isErrorLike } from '../../../../../shared/src/util/errors'
import { queryGraphQL } from '../../../backend/graphql'
import { ProductPlanPrice } from './ProductPlanPrice'
import { ProductPlanTiered } from './ProductPlanTiered'
import { ErrorAlert } from '../../../components/alerts'

interface Props {
    /** The selected plan's billing ID. */
    value: string | null

    /** Called when the selected plan changes (with its billing ID). */
    onChange: (value: string | null) => void

    disabled?: boolean
    className?: string
}

const LOADING: 'loading' = 'loading'

interface State {
    /**
     * The list of all possible product plans.
     */
    plansOrError: GQL.IProductPlan[] | typeof LOADING | ErrorLike
}

/**
 * Displays a form group for selecting a product plan.
 */
export class ProductPlanFormControl extends React.Component<Props, State> {
    public state: State = {
        plansOrError: LOADING,
    }

    private subscriptions = new Subscription()

    public componentDidMount(): void {
        this.subscriptions.add(
            queryProductPlans()
                .pipe(
                    tap(plans => {
                        // If no plan is selected, select the 1st plan when the plans have loaded.
                        if (plans.length > 0 && this.props.value === null) {
                            this.props.onChange(plans[0].billingPlanID)
                        }
                    }),
                    catchError(err => [asError(err)]),
                    startWith(LOADING),
                    map(c => ({ plansOrError: c }))
                )
                .subscribe(stateUpdate => this.setState(stateUpdate))
        )
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        const disableInputs =
            this.props.disabled || this.state.plansOrError === LOADING || isErrorLike(this.state.plansOrError)

        return (
            <div className={`product-plan-form-control ${this.props.className || ''}`}>
                {this.state.plansOrError === LOADING ? (
                    <LoadingSpinner className="icon-inline" />
                ) : isErrorLike(this.state.plansOrError) ? (
                    <ErrorAlert error={this.state.plansOrError.message} />
                ) : (
                    <>
                        <div className="list-group">
                            {this.state.plansOrError.map((plan, i) => (
                                <div key={i} className="list-group-item p-0">
                                    <label className="p-3 mb-0 d-flex" htmlFor={`product-plan-form-control__plan${i}`}>
                                        <input
                                            type="radio"
                                            name="product-plan-form-control__plan"
                                            className="mr-2"
                                            id={`product-plan-form-control__plan${i}`}
                                            value={plan.billingPlanID}
                                            onChange={this.onPlanChange}
                                            required={true}
                                            disabled={disableInputs}
                                            checked={plan.billingPlanID === this.props.value}
                                        />
                                        <div>
                                            <strong>{plan.name}</strong>
                                            <div className="text-muted">
                                                {plan.planTiers.length > 0 ? (
                                                    <ProductPlanTiered
                                                        planTiers={plan.planTiers}
                                                        tierMode={plan.tiersMode}
                                                        minQuantity={plan.minQuantity}
                                                    />
                                                ) : (
                                                    <ProductPlanPrice pricePerUserPerYear={plan.pricePerUserPerYear} />
                                                )}
                                            </div>
                                        </div>
                                    </label>
                                </div>
                            ))}
                        </div>
                        {/* eslint-disable-next-line react/jsx-no-target-blank */}
                        <a href="https://about.sourcegraph.com/pricing" target="_blank" className="small">
                            Compare plans
                        </a>
                    </>
                )}
            </div>
        )
    }

    private onPlanChange: React.ChangeEventHandler<HTMLInputElement> = e => {
        this.props.onChange(e.currentTarget.value)
    }
}

function queryProductPlans(): Observable<GQL.IProductPlan[]> {
    return queryGraphQL(
        gql`
            query ProductPlans {
                dotcom {
                    productPlans {
                        productPlanID
                        billingPlanID
                        name
                        pricePerUserPerYear
                        minQuantity
                        tiersMode
                        planTiers {
                            unitAmount
                            upTo
                            flatAmount
                        }
                    }
                }
            }
        `
    ).pipe(
        map(({ data, errors }) => {
            if (!data || !data.dotcom || !data.dotcom.productPlans || (errors && errors.length > 0)) {
                throw createAggregateError(errors)
            }
            return data.dotcom.productPlans
        })
    )
}
