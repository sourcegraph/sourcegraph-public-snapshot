import React, { useCallback, useEffect } from 'react'

import { mdiMinusThick, mdiPlusThick } from '@mdi/js'
import { useCustomCheckout, PaymentElement, AddressElement } from '@stripe/react-stripe-js'
import classNames from 'classnames'
import { useNavigate } from 'react-router-dom'
import { useDebouncedCallback } from 'use-debounce'

import { pluralize } from '@sourcegraph/common'
import {
    Form,
    Link,
    Button,
    Grid,
    H2,
    Text,
    Container,
    Icon,
    Input,
    Label,
    LoadingSpinner,
    H3,
} from '@sourcegraph/wildcard'

import { CodyAlert } from '../../../components/CodyAlert'

import { PayButton } from './PayButton'

import styles from './NewCodyProSubscriptionPage.module.scss'

export const CodyProCheckoutForm: React.FunctionComponent<{
    creatingTeam: boolean
    customerEmail: string | undefined
}> = ({ creatingTeam, customerEmail }) => {
    const [updatingSeatCount, setUpdatingSeatCount] = React.useState(false)

    const navigate = useNavigate()

    const { total, lineItems, updateLineItemQuantity, email, updateEmail, status } = useCustomCheckout()

    const [errorMessage, setErrorMessage] = React.useState<string | null>(null)
    const [displayErrorMessage, setDisplayErrorMessage] = React.useState(false)
    const [seatCount, setSeatCount] = React.useState(lineItems[0]?.quantity)

    const setAndDisplayErrorMessage = useCallback(
        (message: string) => {
            setErrorMessage(message)
            setDisplayErrorMessage(true)
        },
        [setErrorMessage, setDisplayErrorMessage]
    )

    const debouncedUpdateSeatCount = useDebouncedCallback(async newSeatCount => {
        if (lineItems.length === 1) {
            try {
                await updateLineItemQuantity({
                    lineItem: lineItems[0].id,
                    quantity: newSeatCount,
                })
            } catch(error) {
                setAndDisplayErrorMessage('Failed to update seat count. Please change the number of seats to try again.')
            }
            setUpdatingSeatCount(false)
        }
    }, 800)

    const handlePlusClick = useCallback(
        (diff: number): void => {
            const newSeatCount = Math.min(Math.max(1, seatCount + diff), 50)
            if (newSeatCount === seatCount) {
                return
            }
            setSeatCount(newSeatCount)
            setUpdatingSeatCount(true)
            void debouncedUpdateSeatCount(newSeatCount)
        },
        [debouncedUpdateSeatCount, seatCount]
    )

    useEffect(() => {
        if (lineItems.length === 1) {
            setSeatCount(lineItems[0].quantity)
        }
    }, [lineItems])

    useEffect(() => {
        if (customerEmail) {
            updateEmail(customerEmail)
        }
    }, [customerEmail, updateEmail])

    useEffect(() => {
        if (status.type === 'complete') {
            navigate('/cody/manage?welcome=1')
        } else if (status.type === 'expired') {
            setErrorMessage('Session expired. Please refresh the page.')
        }
    }, [navigate, status.type])

    return (
        <>
            {seatCount >= 30 && (
                <CodyAlert variant="purple">
                    <H3>Explore an enterprise plan</H3>
                    <Text className="mb-0">
                        Team plans are limited to 50 users.{' '}
                        <Link to="https://sourcegraph.com/contact/sales/">Contact sales</Link> to learn more.
                    </Text>
                </CodyAlert>
            )}
            <Container>
                <Grid columnCount={2} spacing={4}>
                    <div>
                        <H2>{creatingTeam ? 'Add seats' : 'Select number of seats'}</H2>
                        <div className="d-flex flex-row align-items-center pb-3 mb-4 border-bottom">
                            <div className="flex-1">$9 per seat / month</div>
                            <Button onClick={() => handlePlusClick(-1)}>
                                <Icon aria-hidden={true} svgPath={mdiMinusThick} />
                            </Button>
                            <div className={styles.seatCountSelectorValue}>{seatCount}</div>
                            <Button onClick={() => handlePlusClick(1)}>
                                <Icon aria-hidden={true} svgPath={mdiPlusThick} />
                            </Button>
                        </div>
                        <H2>Summary</H2>
                        <div className="d-flex flex-row align-items-center mb-4">
                            <div className="flex-1">
                                {creatingTeam ? 'Adding ' : ''} {seatCount} {pluralize('seat', seatCount)}
                            </div>
                            <div>
                                <strong>
                                    {updatingSeatCount ? (
                                        <LoadingSpinner className={styles.lineHeightLoadingSpinner} />
                                    ) : (
                                        `$${total.total / 100} / month`
                                    )}
                                </strong>
                            </div>
                        </div>
                        <Text size="small">
                            <em>Each seat is pro-rated this month, and will be charged at the full rate next month.</em>
                        </Text>
                    </div>
                    <div>
                        <H2>
                            Purchase {seatCount} {pluralize('seat', seatCount)}
                        </H2>
                        <Label>Email</Label>
                        <Input value={email || ''} disabled={true} className="mb-4" />
                        <Form>
                            <PaymentElement options={{ layout: 'accordion' }} className="mb-4" />
                            <AddressElement options={{ mode: 'billing' }} />
                            {errorMessage && displayErrorMessage && (
                                <div className={classNames(styles.paymentDataErrorMessage)}>{errorMessage}</div>
                            )}

                            <PayButton
                                setErrorMessage={setAndDisplayErrorMessage}
                                className={classNames('d-block w-100 mb-4', styles.payButton)}
                            >
                                Subscribe
                            </PayButton>
                            <div>
                                <Text>
                                    By clicking the button, you agree to the{' '}
                                    <Link to="/terms/cloud">Terms of Service</Link> and acknowledge that the{' '}
                                    <Link to="/terms/privacy">Privacy Statement</Link> applies. Your subscription will
                                    renew automatically by charging your payment method on file until you{' '}
                                    <Link to="/docs/cody/usage-and-pricing#downgrading-from-pro-to-free">cancel</Link>.
                                    You may cancel at any time prior to the next billing cycle.
                                </Text>
                            </div>
                        </Form>
                    </div>
                </Grid>
            </Container>
        </>
    )
}
