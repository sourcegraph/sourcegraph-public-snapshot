import * as React from 'react'
import { CardElement, ReactStripeElements } from 'react-stripe-elements'
import { ThemeProps } from '../../../theme'

interface Props extends ThemeProps {
    disabled?: boolean
}

interface State {}

// Workaround for @types/stripe-v3 missing the new disabled attribute. See
// https://github.com/stripe/react-stripe-elements/issues/136#issuecomment-424984951.
type PatchedElementProps = ReactStripeElements.ElementProps & { disabled?: boolean }
const PatchedCardElement = (props: PatchedElementProps) => <CardElement {...props} />

/**
 * Displays a payment form control for the user to enter payment information to purchase a product subscription.
 */
export class PaymentTokenFormControl extends React.Component<Props & ReactStripeElements.InjectedStripeProps, State> {
    public render(): JSX.Element | null {
        const textColor = this.props.isLightTheme ? '#2b3750' : 'white'

        return (
            <div className="payment-token-form-control">
                <PatchedCardElement
                    className={`payment-token-form-control__card payment-token-form-control__card--${
                        this.props.disabled ? 'disabled' : ''
                    }`}
                    disabled={this.props.disabled}
                    // tslint:disable-next-line:jsx-ban-props
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
}
