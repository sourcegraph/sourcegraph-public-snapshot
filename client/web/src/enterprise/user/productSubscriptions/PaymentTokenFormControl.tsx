import { CardElement } from '@stripe/react-stripe-js'
import React from 'react'

import { ThemeProps } from '@sourcegraph/shared/src/theme'

interface Props extends ThemeProps {
    disabled?: boolean
}

/**
 * Displays a payment form control for the user to enter payment information to purchase a product subscription.
 */
export const PaymentTokenFormControl: React.FunctionComponent<Props> = ({ isLightTheme, disabled }) => {
    const textColor = isLightTheme ? '#2b3750' : 'white'
    const bgColor = isLightTheme ? 'white' : '#1d212f'

    return (
        <div className="payment-token-form-control">
            <CardElement
                className={`form-control payment-token-form-control__card payment-token-form-control__card--${
                    disabled ? 'disabled' : ''
                }`}
                options={{
                    disabled,
                    style: {
                        base: {
                            fontFamily:
                                '-apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif',
                            color: textColor,
                            backgroundColor: bgColor,
                            ':-webkit-autofill': {
                                color: textColor,
                            },
                        },
                    },
                }}
            />
        </div>
    )
}
