import React, { useEffect } from 'react'
import { Navigate } from 'react-router-dom'

import classNames from 'classnames'

import { useQuery } from '@sourcegraph/http-client'
import type { TelemetryV2Props } from '@sourcegraph/shared/src/telemetry'
import { PageHeader } from '@sourcegraph/wildcard'

import type { AuthenticatedUser } from '../../../../auth'
import { withAuthenticatedUser } from '../../../../auth/withAuthenticatedUser'

import { Page } from '../../../../components/Page'
import { PageTitle } from '../../../../components/PageTitle'
import { type UserCodyPlanResult, type UserCodyPlanVariables } from '../../../../graphql-operations'
import { CodyProIcon } from '../../../components/CodyIcon'
import { USER_CODY_PLAN } from '../../../subscription/queries'

import styles from './NewCodyProSubscriptionPage.scss'

// NOTE: A side effect of loading this library will update the DOM and
// fetch stripe.js. This is a subtle detail but means that the Stripe
// functionality won't be loaded until this actual module does, via
// the lazily loaded router module.
import * as stripeJs from '@stripe/stripe-js'

import { Elements } from '@stripe/react-stripe-js'
import { CodyProCheckoutForm } from './CodyProCheckoutForm'

const publishableKey = window.context.frontendCodyProConfig?.stripePublishableKey;
if (!publishableKey) {
    console.error("No Stripe publishable key found in config.")
}

// NOTE: Call loadStripe outside a component’s render to avoid recreating the object.
// So here we do it on module load, meaning it gets triggered when the user first
// routes to the page.
const stripePromise =
    stripeJs
        .loadStripe(publishableKey || "")

interface NewCodyProSubscriptionPageProps extends TelemetryV2Props {
    authenticatedUser: AuthenticatedUser
}

// Exported at the bottom of the file, so we can wrap with withAuthenticatedUser.
const newCodyProSubscriptionPage: React.FunctionComponent<NewCodyProSubscriptionPageProps> = ({
    authenticatedUser,
    telemetryRecorder,
}) => {
    useEffect(() => {
        telemetryRecorder.recordEvent('cody.new-subscription-checkout', 'view')
    }, [telemetryRecorder])

    // If the user already has a Cody Pro subscription, direct them back to the Cody Management page.
    const { data, error: dataLoadError } = useQuery<UserCodyPlanResult, UserCodyPlanVariables>(USER_CODY_PLAN, {})
    if (dataLoadError) {
        throw dataLoadError
    }
    if (data?.currentUser?.codySubscription?.plan === 'PRO') {
        return <Navigate to={'/cody/manage'} replace={true} />
    }

    const stripeElementsAppearance: stripeJs.Appearance = {
        theme: 'stripe',
        variables: {
            colorPrimary: '#00b4d9',
        },
    }
    return (
        <>
            <Page className={classNames('d-flex flex-column')}>
                <PageTitle title="New Subscription" />
                <PageHeader className="mb-4 mt-4">
                    <PageHeader.Heading as="h2" styleAs="h1">
                        <div className="d-inline-flex align-items-center">
                            <CodyProIcon className="mr-2" /> Give your team Cody Pro
                        </div>
                    </PageHeader.Heading>
                </PageHeader>

                <div className={classNames('p-4 border bg-1 mt-4', styles.container)}>
                        <Elements stripe={stripePromise} options={{ appearance: stripeElementsAppearance }}>
                            <CodyProCheckoutForm
                                stripeHandle={stripePromise}
                                customerEmail={authenticatedUser?.emails[0].email || ''} />
                        </Elements>
                        {/*
                    <div>
                        <H2>My subscription</H2>
                        <Text className="text-muted mb-0">
                            Hello {authenticatedUser?.displayName}. Depending on whether or not you have a Cody Pro
                            subscription, the Stripe Checkout form or "Manage my Subscription" UI will be displayed
                            here.
                        </Text>

                        <H3>Current Subscription Details</H3>
                        <div>
                            {data?.currentUser?.codySubscription && <pre>{JSON.stringify(data, null, 4)}</pre>}
                        </div>
                    </div>
    */}
                </div>
            </Page>
        </>
    )
}

export const NewCodyProSubscriptionPage = withAuthenticatedUser(newCodyProSubscriptionPage)
