import { useEffect, type FunctionComponent, useMemo } from 'react'

import { Elements } from '@stripe/react-stripe-js'
// NOTE: A side effect of loading this library will update the DOM and
// fetch stripe.js. This is a subtle detail but means that the Stripe
// functionality won't be loaded until this actual module does, via
// the lazily loaded router module.
import { loadStripe } from '@stripe/stripe-js'
import classNames from 'classnames'
import { Navigate, useSearchParams } from 'react-router-dom'

import { useQuery } from '@sourcegraph/http-client'
import type { TelemetryV2Props } from '@sourcegraph/shared/src/telemetry'
import { PageHeader } from '@sourcegraph/wildcard'

import type { AuthenticatedUser } from '../../../../auth'
import { withAuthenticatedUser } from '../../../../auth/withAuthenticatedUser'
import { Page } from '../../../../components/Page'
import { PageTitle } from '../../../../components/PageTitle'
import {
    type UserCodyPlanResult,
    type UserCodyPlanVariables,
    CodySubscriptionPlan,
} from '../../../../graphql-operations'
import { CodyProRoutes } from '../../../codyProRoutes'
import { PageHeaderIcon } from '../../../components/PageHeaderIcon'
import { USER_CODY_PLAN } from '../../../subscription/queries'
import { defaultCodyProApiClientContext, CodyProApiClientContext } from '../../api/components/CodyProApiClient'
import { useBillingAddressStripeElementsOptions } from '../manage/BillingAddress'

import { CodyProCheckoutForm } from './CodyProCheckoutForm'

import styles from './NewCodyProSubscriptionPage.module.scss'

// NOTE: Call loadStripe outside a component’s render to avoid recreating the object.
// We do it here, meaning that "stripe.js" will get loaded lazily, when the user
// routes to this page.
const publishableKey = window.context.frontendCodyProConfig?.stripePublishableKey
const stripe = await loadStripe(publishableKey || '')

interface NewCodyProSubscriptionPageProps extends TelemetryV2Props {
    authenticatedUser: AuthenticatedUser
}

const AuthenticatedNewCodyProSubscriptionPage: FunctionComponent<NewCodyProSubscriptionPageProps> = ({
    authenticatedUser,
    telemetryRecorder,
}) => {
    const [urlSearchParams] = useSearchParams()
    const initialSeatCount = useMemo(() => {
        // Team=1 means the user is purchasing a subscription for a team.
        // Any other value means an individual subscription, except if seats is explicitly set to > 1.
        const defaultSeatCount = urlSearchParams.get('team') === '1' ? 2 : 1
        // If set, we'll use the value set here as the initial seat count.
        const seatCountString = urlSearchParams.get('seats') || defaultSeatCount.toString()
        return parseInt(seatCountString, 10) || defaultSeatCount
    }, [urlSearchParams])
    const isTeam = initialSeatCount > 1

    const options = useBillingAddressStripeElementsOptions()

    useEffect(() => {
        telemetryRecorder.recordEvent('cody.new-subscription-checkout', 'view')
    }, [telemetryRecorder])

    // If the user already has a Cody Pro subscription, direct them back to the Cody Management page.
    const { data, error: dataLoadError } = useQuery<UserCodyPlanResult, UserCodyPlanVariables>(USER_CODY_PLAN, {})
    if (dataLoadError) {
        throw dataLoadError
    }
    if (data?.currentUser?.codySubscription?.plan === CodySubscriptionPlan.PRO) {
        return <Navigate to={CodyProRoutes.Manage} replace={true} />
    }

    return (
        <Page className={classNames('d-flex flex-column', styles.page)}>
            <PageTitle title="New Subscription" />
            <PageHeader className="my-4">
                <PageHeader.Heading as="h1" className={styles.h1}>
                    <div className="d-inline-flex align-items-center">
                        <PageHeaderIcon
                            name={isTeam ? 'mdi-account-multiple-plus-gradient' : 'cody-logo'}
                            className="mr-3"
                        />{' '}
                        {isTeam ? 'Give your team Cody Pro' : 'Upgrade to Cody Pro'}
                    </div>
                </PageHeader.Heading>
            </PageHeader>

            <CodyProApiClientContext.Provider value={defaultCodyProApiClientContext}>
                <Elements stripe={stripe} options={options}>
                    <CodyProCheckoutForm
                        initialSeatCount={initialSeatCount}
                        customerEmail={authenticatedUser?.emails[0].email || ''}
                    />
                </Elements>
            </CodyProApiClientContext.Provider>
        </Page>
    )
}

export const NewCodyProSubscriptionPage = withAuthenticatedUser(AuthenticatedNewCodyProSubscriptionPage)
