import React, { useEffect, type FunctionComponent } from 'react'

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
import { PageHeader, Text, LoadingSpinner, Alert, logger } from '@sourcegraph/wildcard'

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
import { useCurrentSubscription } from '../../api/react-query/subscriptions'
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
    const addSeats = !!urlSearchParams.get('addSeats')
    const isTeam = addSeats || parseInt(urlSearchParams.get('seats') || '', 10) > 1

    const stripeElementsOptions = useBillingAddressStripeElementsOptions()

    // Load data
    const {
        loading: userCodyPlanLoading,
        data: userCodyPlanData,
        error: userCodyPlanError,
    } = useQuery<UserCodyPlanResult, UserCodyPlanVariables>(USER_CODY_PLAN, {})
    const subscriptionQueryResult = useCurrentSubscription()
    const subscription = subscriptionQueryResult?.data

    useEffect(() => {
        telemetryRecorder.recordEvent('cody.new-subscription-checkout', 'view')
    }, [telemetryRecorder])

    useEffect(() => {
        if (userCodyPlanError) {
            logger.error('Failed to fetch subscription data', userCodyPlanError)
        }
    }, [userCodyPlanError])
    useEffect(() => {
        if (subscriptionQueryResult.isError) {
            logger.error('Failed to fetch subscription data', subscriptionQueryResult.error)
        }
    }, [subscriptionQueryResult.isError, subscriptionQueryResult.error])

    if (userCodyPlanLoading || subscriptionQueryResult.isLoading) {
        return <LoadingSpinner className="mx-auto" />
    }

    // If the user already has a Cody Pro subscription, direct them back to the Cody Management page.
    if (!addSeats && userCodyPlanData?.currentUser?.codySubscription?.plan === CodySubscriptionPlan.PRO) {
        return <Navigate to={CodyProRoutes.Manage} replace={true} />
    }

    const PageWithHeader = ({ children }: { children: React.ReactNode }): React.ReactElement => (
        <Page className={classNames('d-flex flex-column', styles.page)}>
            <PageTitle title={addSeats ? 'Add seats' : 'New subscription'} />
            <PageHeader className="my-4 d-inline-flex align-items-center">
                <PageHeader.Heading as="h1" className="text-3xl font-medium">
                    <PageHeaderIcon
                        name={isTeam ? 'mdi-account-multiple-plus-gradient' : 'cody-logo'}
                        className="mr-3"
                    />{' '}
                    <Text as="span">{isTeam ? 'Give your team Cody Pro' : 'Upgrade to Cody Pro'}</Text>
                </PageHeader.Heading>
            </PageHeader>

            {children}
        </Page>
    )

    if (userCodyPlanLoading || subscriptionQueryResult.isLoading) {
        return (
            <PageWithHeader>
                <LoadingSpinner className="mx-auto" />
            </PageWithHeader>
        )
    }

    if (userCodyPlanError) {
        return (
            <PageWithHeader>
                <Alert variant="danger">Failed to fetch user Cody plan data</Alert>
            </PageWithHeader>
        )
    }

    if (addSeats && subscriptionQueryResult.isError) {
        return (
            <PageWithHeader>
                <Alert variant="danger">Failed to fetch subscription data</Alert>
            </PageWithHeader>
        )
    }

    if (addSeats && !subscriptionQueryResult.isLoading && !subscription) {
        return (
            <PageWithHeader>
                <Alert variant="danger">Subscription data is not available</Alert>
            </PageWithHeader>
        )
    }

    return (
        <PageWithHeader>
            <CodyProApiClientContext.Provider value={defaultCodyProApiClientContext}>
                <Elements stripe={stripe} options={stripeElementsOptions}>
                    <CodyProCheckoutForm
                        subscription={subscription}
                        customerEmail={authenticatedUser?.emails[0].email || ''}
                    />
                </Elements>
            </CodyProApiClientContext.Provider>
        </PageWithHeader>
    )
}

export const NewCodyProSubscriptionPage = withAuthenticatedUser(AuthenticatedNewCodyProSubscriptionPage)
