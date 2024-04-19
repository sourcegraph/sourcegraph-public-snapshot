'use client'

import { InvoiceHistoryCard } from './InvoiceHistoryCard.tsx' // TODO: Move it
import { NewSignupCard } from './NewSignupCard.tsx' // TODO: Move it
import { SubscriptionDetailsCard } from './SubscriptionDetailsCard.tsx' // TODO: Move it

import * as stripeJs from '@stripe/stripe-js' // TODO: Export these from @sourcegraph/cody-plg, and import that here
import { Elements } from '@stripe/react-stripe-js' // TODO: Export from @sourcegraph/cody-plg, and import it here
import { type FunctionComponent, useState, useEffect } from 'react'
import { LoadingPage } from './LoadingPage'
import { BackIcon } from './BackIcon'
import { H1, Link, Text } from '@sourcegraph/wildcard'
import { useTemporarySetting } from '@sourcegraph/shared/src/settings/temporary'

// Stripe publishable keys for the test and live environments. These are OK to share externally.
const publishableKeys = {
    test: 'pk_test_0b1ei45h7ypIEeAkXYKBU059',
    live: 'pk_live_1LPIDxv3bZH5wTv9NRcu9Sik',
}

const publishableKey = publishableKeys[location.href.includes('sourcegraph.com') ? 'live' : 'test']
// NOTE: Call loadStripe outside a component’s render to avoid recreating the object.
const stripePromise = stripeJs.loadStripe(publishableKey)

/**
 * Subscription page.
 *
 * Figma Design:
 * https://www.figma.com/file/FMSdn1oKccJRHQPgf7053o/Cody-GA?type=design&node-id=3496-3651&mode=design&t=uuQfSmhMZEoVRBIG-0
 */
export const CodyPlanPage: FunctionComponent = () => {
    // Show the raw GraphQL response for debugging if the local storage key is set.
    const showRawData = useTemporarySetting('sourcegraph.debug.show-subscription-details', false)
    // Show the "new signup" form, even if the user already has a Cody Pro subscription.
    // ☠️ If you complete the flow, the backend will most certainly be in a world of pain. ☠️
    const alwaysShowNewSignup = useTemporarySetting('sourcegraph.debug.always-show-new-signup', false)

    // Load the current user's subscription information.
    // TODO: Replace this with the real back-end communication.
    const data = {currentUser: {
            id: '1',
            name: 'Test User',
            email: 'test@test.com',
            avatarUrl: 'https://avatars.githubusercontent.com/u/1',
            team : {
                id: '1',
                name: 'Test Team',
                subscriptionDetails: {
                    teamId: 'test',
                    primaryEmail: 'test@test.com',
                    name: 'Test subscription',
                    status: 'active',
                },
            }
        }}
    const [loading, setLoading] = useState(true)
    const [error] = useState<Error | null>(null)
    // Set loading to false after two seconds
    useEffect(() => {
        setTimeout(() => {
            setLoading(false)
        }, 2000)
    }, [])

    // Loading indicator.
    if (loading) {
        return <LoadingPage />
    }

    // Error page. We probably should route to a generic 5xx page instead.
    if (error) {
        return (
            <main>
                <H1>Awe snap!</H1>
                <Text>
                    There was an error fetching your subscription information. If this persists, please contact{' '}
                    <Link to="mailto:support@sourcegraph.com">support@sourcegraph.com</Link>
                </Text>
            </main>
        )
    }

    // If the team field is defined, but the subscriptionDetails are not, then we infer the
    // user is a member of a team but does not have the necessary permissions to view the
    // subscription details.
    //
    // BUG: This isn't quite right, as we want to surface some subscription information,
    //      such as if the subscription status is `past_due` or `canceled`. As that would
    //      impact what we display to the user. Until May, 2024 however, we only expect
    //      Sourcegraph team members to see this page. (And that we are adding members to
    //      the official Sourcegraph team manually.)
    const team = data?.currentUser?.team
    const subscriptionDetails = team?.subscriptionDetails
    if (team && !subscriptionDetails) {
        return (
            <main>
                <H1 className="mb-8 mt-20 text-3xl tracking-normal text-slate-900">
                    <img
                        src="/cody/white-box-arrow-bouncing-right.png"
                        className="inline-block mr-2 ssc-icon"
                        alt="Right arrow"
                    />
                    Forbidden
                </H1>

                <p>
                    <a href="/cody/manage" className="flex items-center gap-2 mb-6">
                        <BackIcon />
                        Back to Cody Dashboard
                    </a>
                </p>

                <div className="block container p-6 bg-white border border-separator-gray rounded-lg shadow mb-8">
                    <Text>You do not have permissions to view your Cody Pro subscription information.</Text>
                    <Text>
                        If you believe you are seeing this in error, please contact{' '}
                        <Link to="mailto:support@sourcegraph.com">support@sourcegraph.com</Link>
                    </Text>
                </div>
            </main>
        )
    }

    // If the team field is undefined, the user is not a member of any Cody Pro team and so we prompt them to sign up.
    // (The check for !subscriptionDetails is technically unnecessary, but informs the TS Compiler that after this
    // if-statement that the value is truthy.)
    if (!team || !subscriptionDetails || alwaysShowNewSignup) {
        return (
            <main>
                <H1 className="mb-8 mt-20 text-3xl tracking-normal text-slate-900">
                    <img
                        src="/cody/white-box-arrow-bouncing-right.png"
                        className="inline-block mr-2 ssc-icon"
                        alt="Right arrow"
                    />
                    Upgrade to Cody Pro
                </H1>

                <Text>
                    <Link to="/cody/manage" className="flex items-center gap-2 mb-6">
                        <BackIcon />
                        Back to Cody Dashboard
                    </Link>
                </Text>

                <NewSignupCard stripeHandle={stripePromise} customerEmail={data?.currentUser?.email} />
            </main>
        )
    }

    // If both the team and subscription details are defined, then the user is a team admin
    // and has full access to modify subscription data as needed.
    const appearance: stripeJs.Appearance = {
        theme: 'stripe',
        variables: {
            colorPrimary: '#00b4d9',
        },
    }
    return (
        <Elements stripe={stripePromise} options={{ appearance }}>
            <main>
                {/*
            TODO: Add a top utility bar that matches that of dotcom, to not disorient the user when
            switching to accounts.sourcegraph.com/cody to manage their subscription.
            */}

                <H1 className="mb-8 mt-20 text-3xl tracking-normal text-slate-900">
                    <img
                        src="/cody/white-box-arrow-bouncing-right.png"
                        className="inline-block mr-2 ssc-icon"
                        alt="Right arrow"
                    />
                    Manage Subscription
                </H1>

                <Text>
                    <Link to="/cody/manage" className="flex items-center gap-2 mb-6">
                        <BackIcon />
                        Back to Cody Dashboard
                    </Link>
                </Text>

                <div className="mb-6">
                    <SubscriptionDetailsCard
                        details={subscriptionDetails}
                        refetchSubscription={refetch}
                    />
                </div>
                <InvoiceHistoryCard invoices={subscriptionDetails.invoiceHistory.invoices}/>

                {showRawData && (
                    <div className="container">
                        <h1>Raw Data</h1>
                        <pre className="block whitespace-pre overflow-x-scroll">
                            {JSON.stringify(subscriptionDetails, null, 4)}
                        </pre>
                    </div>
                )}
            </main>
        </Elements>
    )
}
