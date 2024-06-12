import { useEffect } from 'react'

import classNames from 'classnames'
import { Navigate } from 'react-router-dom'

import { logger } from '@sourcegraph/common'
import { useQuery } from '@sourcegraph/http-client'
import type { AuthenticatedUser } from '@sourcegraph/shared/src/auth'
import { Alert, Card, Link, LoadingSpinner, PageHeader, Text } from '@sourcegraph/wildcard'

import { withAuthenticatedUser } from '../../../../auth/withAuthenticatedUser'
import { Page } from '../../../../components/Page'
import { PageTitle } from '../../../../components/PageTitle'
import {
    CodySubscriptionPlan,
    type UserCodyPlanResult,
    type UserCodyPlanVariables,
} from '../../../../graphql-operations'
import type { LegacyLayoutRouteContext } from '../../../../LegacyRouteContext'
import { CodyProRoutes } from '../../../codyProRoutes'
import { USER_CODY_PLAN } from '../../../subscription/queries'
import { useCurrentSubscription } from '../../api/react-query/subscriptions'

import { InvoiceHistory } from './InvoiceHistory'
import { PaymentDetails } from './PaymentDetails'
import { SubscriptionDetails } from './SubscriptionDetails'

import styles from './CodySubscriptionManagePage.module.scss'

interface Props extends Pick<LegacyLayoutRouteContext, 'telemetryRecorder'> {
    authenticatedUser: AuthenticatedUser
}

const AuthenticatedCodySubscriptionManagePage: React.FC<Props> = ({ telemetryRecorder }) => {
    const {
        loading: userCodyPlanLoading,
        error: useCodyPlanError,
        data: userCodyPlanData,
    } = useQuery<UserCodyPlanResult, UserCodyPlanVariables>(USER_CODY_PLAN, {})

    useEffect(
        function recordViewEvent() {
            telemetryRecorder.recordEvent('cody.manage-subscription', 'view')
        },
        [telemetryRecorder]
    )

    if (userCodyPlanLoading) {
        return <LoadingSpinner />
    }

    if (useCodyPlanError) {
        logger.error('Failed to fetch Cody subscription data', useCodyPlanError)
        return null
    }

    const subscriptionData = userCodyPlanData?.currentUser?.codySubscription
    if (!subscriptionData) {
        logger.error('Cody subscription data is not available.')
        return null
    }

    // This page only applies to users who have a Cody Pro subscription to manage.
    // Otherwise, direct them to the ./new page to sign up.
    if (subscriptionData.plan !== CodySubscriptionPlan.PRO) {
        return <Navigate to={CodyProRoutes.NewProSubscription} replace={true} />
    }

    return (
        <Page className="d-flex flex-column">
            <PageContent />
        </Page>
    )
}

const PageContent: React.FC = () => {
    const subscriptionQueryResult = useCurrentSubscription()

    if (subscriptionQueryResult.isLoading) {
        return <LoadingSpinner className="mx-auto" />
    }

    if (subscriptionQueryResult.isError) {
        return <Alert variant="danger">Failed to fetch subscription data</Alert>
    }

    const subscription = subscriptionQueryResult?.data
    if (!subscription) {
        return <Alert variant="warning">Subscription data is not available</Alert>
    }

    return (
        <>
            <PageTitle title="Manage Subscription" />
            <PageHeader className="mt-4">
                <PageHeader.Heading as="h2" styleAs="h1" className="mb-4 d-flex align-items-center">
                    <CodyIcon className="mr-2" />
                    <Text as="span">Manage Subscription</Text>
                </PageHeader.Heading>
            </PageHeader>

            <div className="my-3">
                <Link to={CodyProRoutes.Manage} className="d-flex align-items-center">
                    <BackIcon className="mr-2" />
                    Back to Cody Dashboard
                </Link>
            </div>

            <Card className={classNames('p-4', styles.card)}>
                <SubscriptionDetails subscription={subscription} />

                <hr className={styles.divider} />

                <PaymentDetails subscription={subscription} />
            </Card>

            <Card className={classNames('my-4 p-4', styles.card)}>
                <InvoiceHistory />
            </Card>
        </>
    )
}

const BackIcon: React.FC<{ className?: string }> = props => (
    <svg
        width="16"
        height="16"
        viewBox="0 0 16 16"
        fill="none"
        xmlns="http://www.w3.org/2000/svg"
        className={props.className}
    >
        <title>Back icon</title>
        <path
            d="M6.49967 12.8667L7.43967 11.9267L4.38634 8.86666H15.1663V7.53333H4.38634L7.44634 4.47333L6.49967 3.53333L1.83301 8.19999L6.49967 12.8667Z"
            fill="#0B70DB"
        />
    </svg>
)

/* eslint-disable react/forbid-dom-props */
const CodyIcon: React.FC<{ className?: string }> = ({ className }) => (
    <svg
        width="60"
        height="60"
        viewBox="0 0 60 60"
        fill="none"
        xmlns="http://www.w3.org/2000/svg"
        className={className}
    >
        <path
            d="M0 30C0 20.1711 0 15.2566 1.73454 11.4284C3.68294 7.12819 7.12819 3.68294 11.4284 1.73454C15.2566 0 20.1711 0 30 0C39.8289 0 44.7434 0 48.5716 1.73454C52.8718 3.68294 56.3171 7.12819 58.2655 11.4284C60 15.2566 60 20.1711 60 30C60 39.8289 60 44.7434 58.2655 48.5716C56.3171 52.8718 52.8718 56.3171 48.5716 58.2655C44.7434 60 39.8289 60 30 60C20.1711 60 15.2566 60 11.4284 58.2655C7.12819 56.3171 3.68294 52.8718 1.73454 48.5716C0 44.7434 0 39.8289 0 30Z"
            fill="white"
            style={{ fill: 'white', fillOpacity: 1 }}
        />
        <path
            d="M0 30C0 20.1711 0 15.2566 1.73454 11.4284C3.68294 7.12819 7.12819 3.68294 11.4284 1.73454C15.2566 0 20.1711 0 30 0C39.8289 0 44.7434 0 48.5716 1.73454C52.8718 3.68294 56.3171 7.12819 58.2655 11.4284C60 15.2566 60 20.1711 60 30C60 39.8289 60 44.7434 58.2655 48.5716C56.3171 52.8718 52.8718 56.3171 48.5716 58.2655C44.7434 60 39.8289 60 30 60C20.1711 60 15.2566 60 11.4284 58.2655C7.12819 56.3171 3.68294 52.8718 1.73454 48.5716C0 44.7434 0 39.8289 0 30Z"
            fill="url(#paint0_radial_5107_1290)"
            fillOpacity="0.2"
        />
        <path
            d="M30 59.2C25.0737 59.2 21.4223 59.1995 18.5229 58.9854C15.6314 58.7718 13.5512 58.349 11.7586 57.5368C7.63514 55.6685 4.33153 52.3649 2.46323 48.2414C1.65099 46.4488 1.22819 44.3686 1.01464 41.4771C0.800508 38.5777 0.8 34.9263 0.8 30C0.8 25.0737 0.800508 21.4223 1.01464 18.5229C1.22819 15.6314 1.65099 13.5512 2.46323 11.7586C4.33153 7.63514 7.63514 4.33153 11.7586 2.46323C13.5512 1.65099 15.6314 1.22819 18.5229 1.01464C21.4223 0.800508 25.0737 0.8 30 0.8C34.9263 0.8 38.5777 0.800508 41.4771 1.01464C44.3686 1.22819 46.4488 1.65099 48.2414 2.46323C52.3649 4.33153 55.6685 7.63514 57.5368 11.7586C58.349 13.5512 58.7718 15.6314 58.9854 18.5229C59.1995 21.4223 59.2 25.0737 59.2 30C59.2 34.9263 59.1995 38.5777 58.9854 41.4771C58.7718 44.3686 58.349 46.4488 57.5368 48.2414C55.6685 52.3649 52.3649 55.6685 48.2414 57.5368C46.4488 58.349 44.3686 58.7718 41.4771 58.9854C38.5777 59.1995 34.9263 59.2 30 59.2Z"
            stroke="black"
            strokeOpacity="0.05"
            style={{ stroke: 'black', strokeOpacity: 0.05 }}
            strokeWidth="1.6"
        />
        <path
            fillRule="evenodd"
            clipRule="evenodd"
            d="M36.1019 16.8094C37.5687 16.8094 38.7578 18.0165 38.7578 19.5054L38.7578 25.6675C38.7578 27.1564 37.5687 28.3634 36.1019 28.3634C34.6351 28.3634 33.446 27.1564 33.446 25.6675L33.446 19.5054C33.446 18.0165 34.6351 16.8094 36.1019 16.8094Z"
            fill="url(#paint1_linear_5107_1290)"
        />
        <path
            fillRule="evenodd"
            clipRule="evenodd"
            d="M16.3701 23.7408C16.3701 22.2519 17.5592 21.0449 19.026 21.0449H25.0967C26.5635 21.0449 27.7526 22.2519 27.7526 23.7408C27.7526 25.2298 26.5635 26.4368 25.0967 26.4368H19.026C17.5592 26.4368 16.3701 25.2298 16.3701 23.7408Z"
            fill="url(#paint2_linear_5107_1290)"
        />
        <path
            fillRule="evenodd"
            clipRule="evenodd"
            d="M19.4339 33.4703C18.551 32.2919 16.8952 32.0562 15.7266 32.9458C14.5531 33.8392 14.3153 35.529 15.1954 36.7201C15.202 36.729 15.1961 36.7211 15.1969 36.7222L15.1985 36.7243L15.202 36.729L15.2102 36.7399C15.2161 36.7479 15.2231 36.7571 15.2311 36.7676C15.2472 36.7887 15.2673 36.8148 15.2916 36.8457C15.3403 36.9075 15.4055 36.9884 15.4874 37.0857C15.6511 37.2801 15.8822 37.541 16.1805 37.8457C16.7759 38.4536 17.6485 39.2455 18.7989 40.0333C21.1079 41.6147 24.5566 43.1903 29.0821 43.1903C33.6075 43.1903 37.0562 41.6147 39.3653 40.0333C40.5156 39.2455 41.3883 38.4536 41.9836 37.8457C42.282 37.541 42.513 37.2801 42.6767 37.0857C42.7586 36.9884 42.8239 36.9075 42.8725 36.8457C42.8968 36.8148 42.9496 36.7454 42.9656 36.7243C42.9736 36.7138 42.9687 36.7201 42.954 36.7399L42.9621 36.729L42.9656 36.7243L42.9672 36.7222C42.968 36.7211 42.9621 36.729 42.9656 36.7243L42.9687 36.7201C43.8488 35.529 43.611 33.8392 42.4376 32.9458C41.269 32.0562 39.6132 32.2919 38.7303 33.4702C38.7288 33.4722 38.7261 33.4756 38.7222 33.4805C38.7087 33.4977 38.6808 33.5326 38.6386 33.5828C38.554 33.6833 38.413 33.8437 38.2164 34.0444C37.8217 34.4475 37.2122 35.0035 36.3944 35.5637C34.7669 36.6782 32.3347 37.7985 29.0821 37.7985C25.8294 37.7985 23.3972 36.6782 21.7698 35.5637C20.9519 35.0035 20.3425 34.4475 19.9478 34.0444C19.7511 33.8437 19.6102 33.6833 19.5256 33.5828C19.4834 33.5326 19.4554 33.4977 19.4419 33.4805C19.4381 33.4757 19.4354 33.4722 19.4339 33.4703ZM38.7224 33.4808L38.722 33.4813L38.7207 33.4831C38.72 33.484 38.7193 33.485 38.7193 33.485C38.7203 33.4836 38.7213 33.4822 38.7224 33.4808Z"
            fill="url(#paint3_linear_5107_1290)"
        />
        <defs>
            <radialGradient
                id="paint0_radial_5107_1290"
                cx="0"
                cy="0"
                r="1"
                gradientUnits="userSpaceOnUse"
                gradientTransform="translate(23.4233 -44.4873) rotate(77.074) scale(58.2161)"
            >
                <stop stopColor="#00DBFF" style={{ stopColor: '#00DBFF', stopOpacity: 1 }} />
                <stop offset="1" stopColor="#00DBFF" stopOpacity="0" style={{ stopColor: 'none', stopOpacity: 0 }} />
            </radialGradient>
            <linearGradient
                id="paint1_linear_5107_1290"
                x1="8"
                y1="59.5"
                x2="39"
                y2="17"
                gradientUnits="userSpaceOnUse"
            >
                <stop offset="0.0576364" stopColor="#FF291F" style={{ stopColor: '#FF291F', stopOpacity: 1 }} />
                <stop offset="0.308435" stopColor="#00CBEC" style={{ stopColor: '#00CBEC', stopOpacity: 1 }} />
                <stop offset="0.642062" stopColor="#A112FF" style={{ stopColor: '#A112FF', stopOpacity: 1 }} />
                <stop offset="0.744128" stopColor="#7048E8" style={{ stopColor: '#7048E8', stopOpacity: 1 }} />
                <stop offset="0.876835" stopColor="#FF5543" style={{ stopColor: '#FF5543', stopOpacity: 1 }} />
            </linearGradient>
            <linearGradient
                id="paint2_linear_5107_1290"
                x1="28"
                y1="24"
                x2="18.678"
                y2="26.4367"
                gradientUnits="userSpaceOnUse"
            >
                <stop offset="0.00948703" stopColor="#A112FF" style={{ stopColor: '#A112FF', stopOpacity: 1 }} />
                <stop offset="0.961681" stopColor="#7048E8" style={{ stopColor: '#7048E8', stopOpacity: 1 }} />
            </linearGradient>
            <linearGradient
                id="paint3_linear_5107_1290"
                x1="14.6641"
                y1="27.254"
                x2="53"
                y2="77.5"
                gradientUnits="userSpaceOnUse"
            >
                <stop offset="0.00481802" stopColor="#7048E8" style={{ stopColor: '#7048E8', stopOpacity: 1 }} />
                <stop offset="0.417018" stopColor="#00CBEC" style={{ stopColor: '#00CBEC', stopOpacity: 1 }} />
                <stop offset="0.642062" stopColor="#A112FF" style={{ stopColor: '#A112FF', stopOpacity: 1 }} />
                <stop offset="1" stopColor="#FF5543" style={{ stopColor: '#FF5543', stopOpacity: 1 }} />
            </linearGradient>
        </defs>
    </svg>
)
/* eslint-enable react/forbid-dom-props */

export const CodySubscriptionManagePage = withAuthenticatedUser(AuthenticatedCodySubscriptionManagePage)
