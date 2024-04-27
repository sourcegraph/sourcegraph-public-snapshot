import React, { useEffect } from 'react'

import classNames from 'classnames'

import { useQuery } from '@sourcegraph/http-client'
import type { TelemetryV2Props } from '@sourcegraph/shared/src/telemetry'
import { H2, H3, PageHeader, Text } from '@sourcegraph/wildcard'

import type { AuthenticatedUser } from '../../../auth'
import { Page } from '../../../components/Page'
import { PageTitle } from '../../../components/PageTitle'
import { type UserCodyPlanResult, type UserCodyPlanVariables } from '../../../graphql-operations'
import { CodyProIcon } from '../../components/CodyIcon'
import { USER_CODY_PLAN } from '../../subscription/queries'

import styles from './ManageSubscriptionPage.module.scss'

interface ManageSubscriptionPageProps extends TelemetryV2Props {
    authenticatedUser: AuthenticatedUser | null
}

export const ManageSubscriptionPage: React.FunctionComponent<ManageSubscriptionPageProps> = ({
    authenticatedUser,
    telemetryRecorder,
}) => {
    useEffect(() => {
        telemetryRecorder.recordEvent('cody.new-subscription-checkout', 'view')
    }, [telemetryRecorder])

    const { data, error: dataLoadError } = useQuery<UserCodyPlanResult, UserCodyPlanVariables>(USER_CODY_PLAN, {})
    if (dataLoadError) {
        throw dataLoadError
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
                    <div className="d-flex justify-content-between align-items-center border-bottom pb-3">
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
                    </div>
                </div>
            </Page>
        </>
    )
}
