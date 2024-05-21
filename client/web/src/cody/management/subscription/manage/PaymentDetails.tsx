import { useMemo, useState } from 'react'

import { mdiPencilOutline, mdiCreditCardOutline, mdiPlus } from '@mdi/js'
import {
    AddressElement,
    CardCvcElement,
    CardExpiryElement,
    CardNumberElement,
    Elements,
    useElements,
    useStripe,
} from '@stripe/react-stripe-js'
import { type Appearance, loadStripe, type StripeCardElementOptions } from '@stripe/stripe-js'
import classNames from 'classnames'

import { Button, Form, Grid, H3, Icon, Label, LoadingSpinner, Text } from '@sourcegraph/wildcard'

import type { Subscription } from '../../api/teamSubscriptions'

import styles from './PaymentDetails.module.scss'

const publishableKey = window.context.frontendCodyProConfig?.stripePublishableKey
if (!publishableKey) {
    // TODO: handle error
    throw new Error('Stripe publishable key not found')
}

const stripePromise = loadStripe(publishableKey)

const appearance: Appearance = {
    theme: 'stripe',
    variables: {
        colorPrimary: '#00b4d9',
    },
}

const noop = (): void => {}

export const PaymentDetails: React.FC<{ subscription: Subscription }> = ({ subscription }) => (
    <Elements stripe={stripePromise} options={{ appearance }}>
        <Grid columnCount={2} spacing={0} className={styles.grid}>
            <PaymentMethod subscription={subscription} className={styles.gridItem} onChange={noop} />
            <BillingAddress subscription={subscription} className={styles.gridItem} onChange={noop} />
        </Grid>
    </Elements>
)

const cardElementOptions: StripeCardElementOptions = {
    // Don't use Stripe Link. Just the basics.
    disableLink: true,
    // Since it is supplied by the AddressElement.
    hidePostalCode: true,

    // apply default wildcard input classes
    classes: {
        base: 'form-control',
        focus: 'focus-visible',
        invalid: 'is-invalid',
    },
}

const PaymentMethod: React.FC<{
    subscription: Subscription
    className?: string
    onChange: () => unknown
}> = ({ subscription: { paymentMethod }, className, onChange }) => {
    const [isEditMode, setIsEditMode] = useState(false)

    // // It shouldn't be possible to have a subscription without a payment method.
    // // But this can still happen in some situations.
    // //
    // // For the blank slate experience, we just have a button that toggles the
    // // "editing" state. (After which point, everything will work just like
    // // the "add new payment method" scenario.
    // if (!subscription.paymentMethod && !isEditMode) {
    //     return (
    //         <>
    //             <p>No payment method is available.</p>
    //             <div className="flex justify-end mt-6">
    //                 <button
    //                     type="button"
    //                     className="bg-blue-600 text-white hover:bg-blue-700 py-2 px-4 rounded inline-flex items-center justify-center"
    //                     onClick={() => {
    //                         setEditing(true)
    //                     }}
    //                 >
    //                     Add
    //                 </button>
    //             </div>
    //         </>
    //     )
    // }

    return (
        <div className={className}>
            {paymentMethod ? (
                isEditMode ? (
                    <CreditCardForm onReset={() => setIsEditMode(false)} onSubmit={() => setIsEditMode(false)} />
                ) : (
                    <ActiveCreditCard paymentMethod={paymentMethod} onEditButtonClick={() => setIsEditMode(true)} />
                )
            ) : (
                <CreditCardMissing onAddButtonClick={() => setIsEditMode(true)} />
            )}
        </div>
    )
}

const CreditCardMissing: React.FC<{ onAddButtonClick: () => void }> = props => (
    <div className={styles.creditCardTitle}>
        <H3>No payment method is available</H3>
        <Button variant="link" className={styles.creditCardTitleButton} onClick={props.onAddButtonClick}>
            <Icon aria-hidden={true} svgPath={mdiPlus} className="mr-1" /> Add
        </Button>
    </div>
)

const ActiveCreditCard: React.FC<{
    paymentMethod: Subscription['paymentMethod']
    onEditButtonClick: () => void
}> = props => (
    <>
        <div className={styles.creditCardTitle}>
            <H3>Active credit card</H3>
            <Button variant="link" className={styles.creditCardTitleButton} onClick={props.onEditButtonClick}>
                <Icon aria-hidden={true} svgPath={mdiPencilOutline} className="mr-1" /> Edit
            </Button>
        </div>
        <div className={styles.creditCardContent}>
            <Text as="span" className={classNames('text-muted', styles.creditCardNumber)}>
                <Icon aria-hidden={true} svgPath={mdiCreditCardOutline} /> ···· ···· ···· {props.paymentMethod?.last4}
            </Text>
            <Text as="span" className="text-muted">
                Expires {props.paymentMethod?.expMonth.toString().padStart(2, '0')}/
                {props.paymentMethod?.expYear.toString().slice(-2)}
            </Text>
        </div>
    </>
)

const CreditCardForm: React.FC<{ onReset: () => void; onSubmit: () => void }> = props => {
    const stripe = useStripe()
    const elements = useElements()

    const [isLoading, setIsLoading] = useState(false)
    const [errorMessage, setErrorMessage] = useState('')

    const handleSubmit = async (): Promise<void> => {
        if (!stripe || !elements) {
            return setErrorMessage('Stripe or Stripe Elements libraries not available.')
        }

        const cardNumberElement = elements.getElement(CardNumberElement)
        if (!cardNumberElement) {
            return setErrorMessage('CardNumber element was not found.')
        }

        const tokenResult = await stripe.createToken(cardNumberElement)
        if (tokenResult.error) {
            return setErrorMessage(tokenResult.error.message ?? 'An unknown error occurred.')
        }

        setIsLoading(true)
        try {
            // TODO: call SSC API
            props.onSubmit()
        } catch (error) {
            // TODO[accounts.sourcegraph.com#353]: Send error to Sentry
            // eslint-disable-next-line no-console
            console.error(error)

            // // If there is a human-friendly error in the GraphQL response, surface that to the user.
            // const apolloError = error as ApolloError
            // if (apolloError.name === 'ApolloError') {
            //     if (apolloError.message !== 'Internal Server Error') {
            //         setErrorMessage(
            //             `An error occurred while updating your credit card information: ${apolloError.message}`
            //         )
            //         return
            //     }
            // }
            setErrorMessage(
                'An error occurred while updating your credit card info. Please try again. If the problem persists, contact support at support@sourcegraph.com.'
            )
        } finally {
            setIsLoading(false)
        }
    }

    return (
        <>
            <H3>Edit credit card</H3>
            <Form onSubmit={handleSubmit} onReset={props.onReset} className={styles.creditCardForm}>
                <div>
                    <Label className={styles.creditCardFormLabel}>
                        <Text className="mb-2">Card number</Text>
                        <CardNumberElement options={cardElementOptions} onFocus={() => {}} />
                    </Label>
                </div>

                <Grid columnCount={2} className="mt-3 mb-0 pb-3">
                    <Label className={styles.creditCardFormLabel}>
                        <Text className="mb-2">Expiry date</Text>
                        <CardExpiryElement options={cardElementOptions} onFocus={() => {}} />
                    </Label>

                    <Label className={styles.creditCardFormLabel}>
                        <Text className="mb-2">CVC</Text>
                        <CardCvcElement options={cardElementOptions} onFocus={() => {}} />
                    </Label>
                </Grid>

                {errorMessage && <Text className="text-danger">{errorMessage}</Text>}

                <div className={classNames('mt-4', styles.creditCardFormButtonContainer)}>
                    <Button type="reset" variant="secondary" outline={true}>
                        Cancel
                    </Button>
                    <Button disabled={isLoading} type="submit" variant="primary" className="ml-2">
                        Save
                    </Button>
                </div>
            </Form>
        </>
    )
}

const BillingAddress: React.FC<{
    subscription: Subscription
    className?: string
    onChange: () => unknown
}> = ({ subscription, className, onChange }) => {
    const stripe = useStripe()
    const elements = useElements()
    const [editing, setEditing] = useState(false)
    const [errorMessage, setErrorMessage] = useState('')
    const [savingStatus, setSavingStatus] = useState(false)

    const handleSubmit = async () => {
        if (!stripe || !elements) {
            setErrorMessage('Stripe or Stripe Elements libraries not available.')
            return
        }
        const addressElement = elements.getElement(AddressElement)
        if (!addressElement) {
            setErrorMessage('AddressElement not found.')
            return
        }
        const addressElementValue = await addressElement.getValue()
        if (!addressElementValue.complete) {
            setErrorMessage('Address is not complete.')
            return
        }
        const suppliedAddress = addressElementValue.value.address

        setSavingStatus(true)
        try {
            const addressChanged =
                suppliedAddress.line1 !== subscription.address.line1 ||
                suppliedAddress.line2 !== subscription.address.line2 ||
                suppliedAddress.city !== subscription.address.city ||
                suppliedAddress.state !== subscription.address.state ||
                suppliedAddress.postal_code !== subscription.address.postalCode ||
                suppliedAddress.country !== subscription.address.country
            const nameChanged = addressElementValue.value.name !== subscription.name

            if (addressChanged) {
                // await client.mutate({
                //     mutation: MUTATE_TEAM_SUBSCRIPTION_ADDRESS,
                //     variables: {
                //         teamId: subscription.teamId,
                //         addressLine1: suppliedAddress.line1,
                //         addressLine2: suppliedAddress.line2 ?? '',
                //         addressCity: suppliedAddress.city,
                //         addressState: suppliedAddress.state,
                //         addressPostalCode: suppliedAddress.postal_code,
                //         addressCountry: suppliedAddress.country,
                //     },
                // })

                console.log('TODO: mutate team subscription address')
            }

            if (nameChanged) {
                // await client.mutate({
                //     mutation: MUTATE_TEAM_SUBSCRIPTION_CUSTOMER_NAME,
                //     variables: {
                //         teamId: subscription.teamId,
                //         customerName: addressElementValue.value.name,
                //     },
                // })

                console.log('TODO: mutate team subscription customer name')
            }

            if (addressChanged || nameChanged) {
                onChange()
            }
        } catch (error) {
            // TODO[accounts.sourcegraph.com#353]: Send error to Sentry
            // eslint-disable-next-line no-console
            console.error(error)
            setErrorMessage(
                'An error occurred while updating your contact information. Please try again. If the problem persists, contact support at support@sourcegraph.com.'
            )
        }
        setSavingStatus(false)
        setEditing(false)
    }

    // TODO: Customize this further, enabling validation, default forms, autocomplete, etc.
    // https://stripe.com/docs/js/elements_object/create_address_element
    const options = useMemo(
        () => ({
            mode: 'billing',
            display: { name: 'full' },
            defaultValues: {
                name: subscription.name,
                address: {
                    line1: subscription.address.line1,
                    line2: subscription.address.line2,
                    city: subscription.address.city,
                    state: subscription.address.state,
                    postal_code: subscription.address.postalCode,
                    country: subscription.address.country,
                },
            },
        }),
        [subscription]
    )

    return (
        <div className={className}>
            {!editing && (
                <a
                    className="float-right cursor-pointer inline-flex items-center"
                    onClick={() => {
                        setEditing(true)
                    }}
                >
                    <svg
                        width="13"
                        height="13"
                        viewBox="0 0 13 13"
                        fill="none"
                        xmlns="http://www.w3.org/2000/svg"
                        className="mr-2"
                    >
                        <path
                            d="M7.87333 4.20001L8.5 4.82668L2.44667 10.8667H1.83333V10.2533L7.87333 4.20001ZM10.2733 0.200012C10.1067 0.200012 9.93333 0.266679 9.80667 0.393346L8.58667 1.61335L11.0867 4.11335L12.3067 2.89335C12.5667 2.63335 12.5667 2.20001 12.3067 1.95335L10.7467 0.393346C10.6133 0.260012 10.4467 0.200012 10.2733 0.200012ZM7.87333 2.32668L0.5 9.70001V12.2H3L10.3733 4.82668L7.87333 2.32668Z"
                            fill="#0B70DB"
                        />
                    </svg>
                    Edit
                </a>
            )}
            <h2 className="text-lg font-semibold mb-6 text-slate-950">Billing address</h2>
            {errorMessage && <p className="bg-red-100 text-red-700 p-4 rounded">{errorMessage}</p>}

            <div className="relative">
                {editing ? (
                    <AddressElement
                        options={options}
                        onFocus={() => {
                            setErrorMessage(null)
                        }}
                    />
                ) : (
                    <div>
                        <p className="m-0 text-xs text-muted">Full name</p>
                        <p className="m-0 mb-4">{subscription.name}</p>

                        <p className="m-0 text-xs text-muted">Country or region</p>
                        <p className="m-0 mb-4">{subscription.address.country || '-'}</p>

                        <p className="m-0 text-xs text-muted">Address line 1</p>
                        <p className="m-0 mb-4">{subscription.address.line1 || '-'}</p>

                        <p className="m-0 text-xs text-muted">Address line 2</p>
                        <p className="m-0 mb-4">{subscription.address.line2 || '-'}</p>

                        <p className="m-0 text-xs text-muted">City</p>
                        <p className="m-0 mb-4">{subscription.address.city || '-'}</p>

                        <p className="m-0 text-xs text-muted">State</p>
                        <p className="m-0 mb-4">{subscription.address.state || '-'}</p>

                        <p className="m-0 text-xs text-muted">Postal code</p>
                        <p className="m-0 mb-4">{subscription.address.postalCode || '-'}</p>
                    </div>
                )}
            </div>

            <div className="flex justify-end mt-6">
                {editing && (
                    <>
                        <button
                            type="button"
                            className="bg-gray-200 text-gray-700 hover:bg-gray-300 py-2 px-4 rounded mr-2"
                            onClick={() => {
                                setEditing(false)
                            }}
                        >
                            Cancel
                        </button>
                        <button
                            type="button"
                            className="bg-blue-600 text-white hover:bg-blue-700 py-2 px-4 rounded inline-flex items-center justify-center"
                            onClick={event => {
                                event.preventDefault()
                                void handleSubmit()
                            }}
                            disabled={savingStatus}
                        >
                            {savingStatus && (
                                <div className="spinner w-4 h-4 border-2 border-blue-700 border-t-transparent rounded-full animate-spin mr-2" />
                            )}
                            Submit
                        </button>
                    </>
                )}
            </div>
        </div>
    )
}
