import * as React from 'react'

import classNames from 'classnames'
import { CardElement, ReactStripeElements } from 'react-stripe-elements'

import { ThemeProps } from '@sourcegraph/shared/src/theme'

import styles from './PaymentTokenFormControl.module.scss'

interface Props extends ThemeProps {
    disabled?: boolean
}

// Workaround for @types/stripe-v3 missing the new disabled attribute. See
// https://github.com/stripe/react-stripe-elements/issues/136#issuecomment-424984951.
type PatchedElementProps = ReactStripeElements.ElementProps & { disabled?: boolean }
const PatchedCardElement: React.FunctionComponent<React.PropsWithChildren<PatchedElementProps>> = props => (
    <CardElement {...props} />
)

/**
 * Displays a payment form control for the user to enter payment information to purchase a product subscription.
 */
export const PaymentTokenFormControl: React.FunctionComponent<React.PropsWithChildren<Props>> = props => {
    const textColor = props.isLightTheme ? '#2b3750' : 'white'

    return (
        <div className="payment-token-form-control">
            <PatchedCardElement
                className={classNames('form-control', styles.card, props.disabled && styles.cardDisabled)}
                disabled={props.disabled}
                style={{
                    base: {
                        fontFamily:
                            '-apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif',
                        color: textColor,
                        ':-webkit-autofill': {
                            color: textColor,
                        },
                    },
                }}
            />
        </div>
    )
}
