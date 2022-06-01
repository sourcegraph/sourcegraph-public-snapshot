import React, { useCallback } from 'react'

import classNames from 'classnames'

import { Typography, Input } from '@sourcegraph/wildcard'

interface Props {
    /** The user count input by the user. */
    value: number | null

    /** Called when the user count value changes. */
    onChange: (value: number | null) => void

    disabled?: boolean
    className?: string
}

/**
 * The minimum user count.
 */
export const MIN_USER_COUNT = 25

/**
 * The step size by which to increment/decrement user count.
 */
const USER_COUNT_STEP = 25

/**
 * Displays a form control for inputting the user count for a product subscription.
 */
export const ProductSubscriptionUserCountFormControl: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
    value,
    onChange,
    disabled,
    className = '',
}) => {
    const onUserCountChange = useCallback<React.ChangeEventHandler<HTMLInputElement>>(
        event => {
            // Check for NaN (which is the value if the user deletes the input's value).
            onChange(Number.isNaN(event.currentTarget.valueAsNumber) ? null : event.currentTarget.valueAsNumber)
        },
        [onChange]
    )

    return (
        <div className={classNames('product-subscription-user-count-control form-group align-items-center', className)}>
            <Typography.Label
                htmlFor="product-subscription-user-count-control__userCount"
                className="mb-0 mr-2"
                weight="bold"
            >
                Users
            </Typography.Label>
            <div className="d-flex align-items-center">
                <Input
                    id="product-subscription-user-count-control__userCount"
                    type="number"
                    className="w-auto"
                    min={MIN_USER_COUNT}
                    step={USER_COUNT_STEP}
                    max={50000}
                    required={true}
                    disabled={disabled}
                    value={value || ''}
                    onChange={onUserCountChange}
                />
            </div>
        </div>
    )
}
