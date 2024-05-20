import { useEffect, useMemo } from 'react'

import classNames from 'classnames'
import { Navigate } from 'react-router-dom'

import { useQuery } from '@sourcegraph/http-client'
import type { AuthenticatedUser } from '@sourcegraph/shared/src/auth'
import { lazyComponent } from '@sourcegraph/shared/src/util/lazyComponent'
import { Card, Link, LoadingSpinner, PageHeader, Text } from '@sourcegraph/wildcard'

import { withAuthenticatedUser } from '../../../../auth/withAuthenticatedUser'
import { Page } from '../../../../components/Page'
import { PageTitle } from '../../../../components/PageTitle'
import {
    CodySubscriptionPlan,
    type UserCodyPlanResult,
    type UserCodyPlanVariables,
} from '../../../../graphql-operations'
import type { LegacyLayoutRouteContext } from '../../../../LegacyRouteContext'
import { BackIcon, DashboardIcon } from '../../../components/CodyIcon'
import { USER_CODY_PLAN } from '../../../subscription/queries'
import { Client } from '../../api/client'
import { useApiCaller } from '../../api/hooks/useApiClient'

import { InvoiceHistory } from './InvoiceHistory'
import { SubscriptionDetails } from './SubscriptionDetails'

import styles from './CodySubscriptionManagePage.module.scss'

const PaymentDetails = lazyComponent(() => import('./PaymentDetails'), 'PaymentDetails')

interface Props extends Pick<LegacyLayoutRouteContext, 'telemetryRecorder'> {
    authenticatedUser: AuthenticatedUser
}

const AuthenticatedCodySubscriptionManagePage: React.FC<Props> = ({ telemetryRecorder }) => {
    const {
        loading: userCodyPlanLoading,
        error: useCodyPlanError,
        data: userCodyPlanData,
    } = useQuery<UserCodyPlanResult, UserCodyPlanVariables>(USER_CODY_PLAN, {})

    useEffect(() => {
        telemetryRecorder.recordEvent('cody.manage-subscription', 'view')
    }, [telemetryRecorder])

    if (userCodyPlanLoading) {
        return <LoadingSpinner />
    }

    if (useCodyPlanError) {
        // TODO: handle error
        return null
    }

    const subscriptionData = userCodyPlanData?.currentUser?.codySubscription
    if (!subscriptionData) {
        // TODO: why empty response - handle it
        return null
    }

    // This page only applies to users who have a Cody Pro subscription to manage.
    // Otherwise, direct them to the ./new page to sign up.
    if (subscriptionData.plan !== CodySubscriptionPlan.PRO) {
        return <Navigate to="/cody/manage/subscription/new" replace={true} />
    }

    return <PageContent />
}

const currentSubscriptionCall = Client.getCurrentSubscription()

const PageContent: React.FC = () => {
    const { loading, error, data } = useApiCaller(currentSubscriptionCall)
    // TODO: remove mock data usage!
    const subscription = useMemo(
        () =>
            data && {
                ...data,
                address: {
                    line1: '742 Evergreen Terrace',
                    line2: '',
                    city: 'Springfield',
                    state: 'IL',
                    postalCode: '62629',
                    country: 'US',
                },
                paymentMethod: {
                    expMonth: 6,
                    expYear: 30,
                    last4: '4242',
                },
            },
        [data]
    )
    if (loading) {
        return <LoadingSpinner />
    }

    if (error) {
        // TODO: handle error
        return null
    }

    if (!subscription) {
        // TODO: why empty response - handle it
        return null
    }

    return (
        <Page className={classNames('d-flex flex-column')}>
            <PageTitle title="Manage Subscription" />
            <PageHeader className="mt-4">
                <PageHeader.Heading as="h2" styleAs="h1" className="mb-4">
                    <div className="mb-3">
                        <DashboardIcon className="mr-3" />
                        Manage Subscription
                    </div>
                </PageHeader.Heading>
            </PageHeader>

            <div className="my-4">
                <Link to="/cody/manage" className={styles.link}>
                    <BackIcon />
                    Back to Cody Dashboard
                </Link>
            </div>

            <Card className={classNames('p-4', styles.card)}>
                <div className="mb-3">
                    <SubscriptionDetails subscription={subscription} />
                </div>

                <hr className="w-100 my-2" />

                <PaymentDetails subscription={subscription} />
            </Card>

            <InvoiceHistory />
        </Page>
    )
}

export const CodySubscriptionManagePage = withAuthenticatedUser(AuthenticatedCodySubscriptionManagePage)
