import * as GQL from '@sourcegraph/webapp/dist/backend/graphqlschema'
import { numberWithCommas } from '@sourcegraph/webapp/dist/util/strings'
import * as React from 'react'
import { ReactStripeElements } from 'react-stripe-elements'
import { ProductSubscriptionInput } from '../../dotcom/productSubscriptions/helpers'
import { formatUserCount } from '../../productSubscription/helpers'
import { PaymentTokenFormControl } from './PaymentTokenFormControl'

interface Props {
    user: GQL.IUser
    productSubscription: ProductSubscriptionInput | null
    disabled?: boolean
    isLightTheme: boolean
}

/**
 * Displays the payment section of the new product subscription form.
 */
export const NewProductSubscriptionPaymentSection: React.SFC<
    Props & ReactStripeElements.InjectedStripeProps
> = props => (
    <div className="new-product-subscription-payment-section">
        <div className="form-text mb-2">
            Total:{' '}
            {props.productSubscription ? (
                `$${numberWithCommas(props.productSubscription.totalPriceNonAuthoritative / 100)}`
            ) : (
                <>&mdash;</>
            )}{' '}
            for 1 year {props.productSubscription && <>({formatUserCount(props.productSubscription.userCount)})</>}
        </div>
        <PaymentTokenFormControl disabled={props.disabled} isLightTheme={props.isLightTheme} />
    </div>
)
