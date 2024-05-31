import { useEffect, type FunctionComponent } from 'react'

import { Elements } from '@stripe/react-stripe-js'
// NOTE: A side effect of loading this library will update the DOM and
// fetch stripe.js. This is a subtle detail but means that the Stripe
// functionality won't be loaded until this actual module does, via
// the lazily loaded router module.
import * as stripeJs from '@stripe/stripe-js'
import classNames from 'classnames'
import { Navigate, useSearchParams } from 'react-router-dom'

import { useQuery } from '@sourcegraph/http-client'
import type { TelemetryV2Props } from '@sourcegraph/shared/src/telemetry'
import { Container, PageHeader } from '@sourcegraph/wildcard'

import type { AuthenticatedUser } from '../../../../auth'
import { withAuthenticatedUser } from '../../../../auth/withAuthenticatedUser'
import { Page } from '../../../../components/Page'
import { PageTitle } from '../../../../components/PageTitle'
import {
    type UserCodyPlanResult,
    type UserCodyPlanVariables,
    CodySubscriptionPlan,
} from '../../../../graphql-operations'
import { WhiteIcon } from '../../../components/WhiteIcon'
import { USER_CODY_PLAN } from '../../../subscription/queries'
import { defaultCodyProApiClientContext, CodyProApiClientContext } from '../../api/components/CodyProApiClient'

import { CodyProCheckoutForm } from './CodyProCheckoutForm'

// NOTE: Call loadStripe outside a component’s render to avoid recreating the object.
// We do it here, meaning that "stripe.js" will get loaded lazily, when the user
// routes to this page.
const publishableKey = window.context.frontendCodyProConfig?.stripePublishableKey
const stripePromise = stripeJs.loadStripe(publishableKey || '')

interface NewCodyProSubscriptionPageProps extends TelemetryV2Props {
    authenticatedUser: AuthenticatedUser
}

const AuthenticatedNewCodyProSubscriptionPage: FunctionComponent<NewCodyProSubscriptionPageProps> = ({
    authenticatedUser,
    telemetryRecorder,
}) => {
    const [urlSearchParams] = useSearchParams()
    const creatingTeam = urlSearchParams.get('team') === '1'

    useEffect(() => {
        telemetryRecorder.recordEvent('cody.new-subscription-checkout', 'view')
    }, [telemetryRecorder])

    // If the user already has a Cody Pro subscription, direct them back to the Cody Management page.
    const { data, error: dataLoadError } = useQuery<UserCodyPlanResult, UserCodyPlanVariables>(USER_CODY_PLAN, {})
    if (dataLoadError) {
        throw dataLoadError
    }
    if (data?.currentUser?.codySubscription?.plan === CodySubscriptionPlan.PRO) {
        return <Navigate to="/cody/manage" replace={true} />
    }

    const stripeElementsAppearance: stripeJs.Appearance = {
        theme: 'stripe',
        variables: {
            colorPrimary: '#00b4d9',
        },
    }
    return (
        <Page className={classNames('d-flex flex-column')}>
            <PageTitle title="New Subscription" />
            <PageHeader className="mb-4 mt-4">
                <PageHeader.Heading as="h2" styleAs="h1">
                    <div className="d-inline-flex align-items-center">
                        <WhiteIcon name="cody-logo" className="mr-3" />{' '}
                        {creatingTeam ? 'Give your team Cody Pro' : 'Upgrade to Cody Pro'}
                    </div>
                </PageHeader.Heading>
            </PageHeader>

            <Container>
                <CodyProApiClientContext.Provider value={defaultCodyProApiClientContext}>
                    <Elements stripe={stripePromise} options={{ appearance: stripeElementsAppearance }}>
                        <CodyProCheckoutForm
                            stripePromise={stripePromise}
                            customerEmail={authenticatedUser?.emails[0].email || ''}
                        />
                    </Elements>
                </CodyProApiClientContext.Provider>
            </Container>
        </Page>
    )
}

export const NewCodyProSubscriptionPage = withAuthenticatedUser(AuthenticatedNewCodyProSubscriptionPage)
