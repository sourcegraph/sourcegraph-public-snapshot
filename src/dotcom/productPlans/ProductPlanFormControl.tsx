import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { gql, queryGraphQL } from '@sourcegraph/webapp/dist/backend/graphql'
import * as GQL from '@sourcegraph/webapp/dist/backend/graphqlschema'
import { asError, createAggregateError, ErrorLike, isErrorLike } from '@sourcegraph/webapp/dist/util/errors'
import { upperFirst } from 'lodash'
import * as React from 'react'
import { Observable, Subscription } from 'rxjs'
import { catchError, map, startWith } from 'rxjs/operators'
import { ProductPlanPrice } from './ProductPlanPrice'

interface Props {
    /** The selected plan. */
    value: GQL.IProductPlan | null

    /** Called when the selected plan changes. */
    onChange: (value: GQL.IProductPlan | null) => void

    disabled?: boolean
    className?: string
}

const LOADING: 'loading' = 'loading'

interface State {
    /** The selected plan. */
    plan: GQL.IProductPlan | null

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
        plan: null,
        plansOrError: LOADING,
    }

    private subscriptions = new Subscription()

    public componentDidMount(): void {
        this.subscriptions.add(
            queryProductPlans()
                .pipe(
                    catchError(err => [asError(err)]),
                    startWith(LOADING),
                    map(c => ({ plansOrError: c }))
                )
                .subscribe(stateUpdate => this.setState(stateUpdate))
        )
    }

    public componentDidUpdate(_prevProps: Props, prevState: State): void {
        if (this.state.plansOrError !== prevState.plansOrError) {
            // Set the default values for the inputs when the data is available.
            let effectivePlan: GQL.IProductPlan | undefined
            if (this.state.plansOrError !== LOADING && !isErrorLike(this.state.plansOrError)) {
                if (this.state.plan) {
                    effectivePlan = this.state.plansOrError.find(plan => this.state.plan === plan)
                } else if (this.state.plansOrError.length > 0) {
                    // Default to first plan.
                    effectivePlan = this.state.plansOrError[0]
                }
            }

            this.setState({ plan: effectivePlan || null }, () => this.props.onChange(this.state.plan))
        }
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
                    <div className="alert alert-danger">{upperFirst(this.state.plansOrError.message)}</div>
                ) : (
                    this.state.plan && (
                        <>
                            <div className="list-group">
                                {this.state.plansOrError.map((plan, i) => (
                                    <div key={i} className="list-group-item p-0">
                                        <label
                                            className="p-3 mb-0 d-flex"
                                            htmlFor={`product-plan-form-control__plan${i}`}
                                        >
                                            <input
                                                type="radio"
                                                name="product-plan-form-control__plan"
                                                className="mr-2"
                                                id={`product-plan-form-control__plan${i}`}
                                                value={plan.billingPlanID}
                                                onChange={this.onPlanChange}
                                                required={true}
                                                disabled={disableInputs}
                                                checked={plan === this.state.plan}
                                            />
                                            <div>
                                                <strong>{plan.name}</strong>
                                                <div className="text-muted">
                                                    <ProductPlanPrice pricePerUserPerYear={plan.pricePerUserPerYear} />
                                                </div>
                                            </div>
                                        </label>
                                    </div>
                                ))}
                            </div>
                            <a href="https://about.sourcegraph.com/pricing" target="_blank" className="small">
                                Compare plans
                            </a>
                        </>
                    )
                )}
            </div>
        )
    }

    private onPlanChange: React.ChangeEventHandler<HTMLInputElement> = e => {
        const value = e.currentTarget.value
        this.setState(
            prevState => ({
                plan:
                    (prevState.plansOrError !== LOADING &&
                        !isErrorLike(prevState.plansOrError) &&
                        prevState.plansOrError.find(plan => plan.billingPlanID === value)) ||
                    null,
            }),
            () => this.props.onChange(this.state.plan)
        )
    }
}

function queryProductPlans(): Observable<GQL.IProductPlan[]> {
    return queryGraphQL(
        gql`
            query ProductPlans {
                dotcom {
                    productPlans {
                        billingPlanID
                        name
                        pricePerUserPerYear
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
