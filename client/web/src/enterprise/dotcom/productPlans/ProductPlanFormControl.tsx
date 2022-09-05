import React, { useCallback, useMemo } from 'react'

import classNames from 'classnames'
import { Observable } from 'rxjs'
import { catchError, map, startWith, tap } from 'rxjs/operators'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { asError, createAggregateError, isErrorLike } from '@sourcegraph/common'
import { gql } from '@sourcegraph/http-client'
import * as GQL from '@sourcegraph/shared/src/schema'
import { RadioButton, LoadingSpinner, useObservable, Link } from '@sourcegraph/wildcard'

import { queryGraphQL } from '../../../backend/graphql'

import { ProductPlanPrice } from './ProductPlanPrice'

interface Props {
    /** The selected plan's billing ID. */
    value: string | null

    /** Called when the selected plan changes (with its billing ID). */
    onChange: (value: string | null) => void

    disabled?: boolean
    className?: string

    /** For mocking in tests only. */
    _queryProductPlans?: typeof queryProductPlans
}

const LOADING = 'loading' as const

/**
 * Displays a form group for selecting a product plan.
 */
export const ProductPlanFormControl: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
    value,
    onChange,
    disabled,
    className = '',
    _queryProductPlans = queryProductPlans,
}) => {
    const noPlanSelected = value === null // don't recompute observable below on every value change

    /**
     * The list of all possible product plans, loading, or an error.
     */
    const plans =
        useObservable(
            useMemo(
                () =>
                    _queryProductPlans().pipe(
                        tap(plans => {
                            // If no plan is selected, select the 1st plan when the plans have loaded.
                            if (plans.length > 0 && noPlanSelected) {
                                onChange(plans[0].billingPlanID)
                            }
                        }),
                        catchError(error => [asError(error)]),
                        startWith(LOADING)
                    ),
                [_queryProductPlans, onChange, noPlanSelected]
            )
        ) || LOADING

    const onPlanChange = useCallback<React.ChangeEventHandler<HTMLInputElement>>(
        event => {
            onChange(event.currentTarget.value)
        },
        [onChange]
    )

    const disableInputs = disabled || plans === LOADING || isErrorLike(plans)

    return (
        <div className={classNames('product-plan-form-control', className)}>
            {plans === LOADING ? (
                <LoadingSpinner />
            ) : isErrorLike(plans) ? (
                <ErrorAlert error={plans.message} />
            ) : (
                <>
                    <div className="list-group">
                        {plans.map((plan, index) => (
                            <div key={plan.billingPlanID} className="list-group-item p-0">
                                <div className="p-3 mb-0 d-flex">
                                    <RadioButton
                                        name="product-plan-form-control__plan"
                                        className="mr-2"
                                        id={`product-plan-form-control__plan${index}`}
                                        value={plan.billingPlanID}
                                        onChange={onPlanChange}
                                        required={true}
                                        disabled={disableInputs}
                                        checked={plan.billingPlanID === value}
                                        label={
                                            <div>
                                                <strong>{plan.name}</strong>
                                                <div className="text-muted">
                                                    <ProductPlanPrice plan={plan} />
                                                </div>
                                            </div>
                                        }
                                    />
                                </div>
                            </div>
                        ))}
                    </div>
                    <Link to="https://about.sourcegraph.com/pricing" className="small">
                        Compare plans
                    </Link>
                </>
            )}
        </div>
    )
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
                        maxQuantity
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
