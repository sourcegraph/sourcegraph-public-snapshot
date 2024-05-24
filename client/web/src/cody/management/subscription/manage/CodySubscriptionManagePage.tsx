import { useEffect } from 'react'

import classNames from 'classnames'
import { Navigate } from 'react-router-dom'

import { logger } from '@sourcegraph/common'
import { useQuery } from '@sourcegraph/http-client'
import type { AuthenticatedUser } from '@sourcegraph/shared/src/auth'
import { lazyComponent } from '@sourcegraph/shared/src/util/lazyComponent'
import { Card, Link, LoadingSpinner, PageHeader } from '@sourcegraph/wildcard'

import { withAuthenticatedUser } from '../../../../auth/withAuthenticatedUser'
import { Page } from '../../../../components/Page'
import { PageTitle } from '../../../../components/PageTitle'
import {
    CodySubscriptionPlan,
    type UserCodyPlanResult,
    type UserCodyPlanVariables,
} from '../../../../graphql-operations'
import type { LegacyLayoutRouteContext } from '../../../../LegacyRouteContext'
import { BackIcon } from '../../../components/CodyIcon'
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
        return <Navigate to="/cody/manage/subscription/new" replace={true} />
    }

    return <PageContent />
}

const currentSubscriptionCall = Client.getCurrentSubscription()

const PageContent: React.FC = () => {
    const {
        loading,
        error,
        data: subscription,
        refetch: refetchSubscription,
        response,
    } = useApiCaller(currentSubscriptionCall)

    if (loading) {
        return <LoadingSpinner />
    }

    if (error) {
        logger.error('Error fetching current subscription', error)
        return null
    }

    if (response && !response.ok) {
        if (response.status === 401) {
            return <Navigate to="/-/sign-out" replace={true} />
        }

        logger.error(`Fetch Cody subscription request failed with status ${response.status}`)
        return null
    }

    if (!subscription) {
        if (response) {
            logger.error('Current subscription is not available.')
        }
        return null
    }

    return (
        <Page className="d-flex flex-column">
            <PageTitle title="Manage Subscription" />
            <PageHeader className="mt-4">
                <PageHeader.Heading as="h2" styleAs="h1" className="mb-4">
                    <div className="mb-3">
                        <CodyIcon className={styles.codyIcon} />
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

            <Card className={classNames('my-4 p-4', styles.card)}>
                <SubscriptionDetails subscription={subscription} refetchSubscription={refetchSubscription} />

                <hr className={styles.divider} />

                <PaymentDetails subscription={subscription} refetchSubscription={refetchSubscription} />
            </Card>

            <Card className={classNames('my-4 p-4', styles.card)}>
                <InvoiceHistory />
            </Card>
        </Page>
    )
}

const CodyIcon: React.FC<{ className?: string }> = ({ className }) => (
    <svg
        width="92"
        height="92"
        viewBox="0 0 92 92"
        fill="none"
        xmlns="http://www.w3.org/2000/svg"
        className={className}
    >
        <g filter="url(#filter0_dd_3496_6128)">
            <path
                d="M16.5 39.6C16.5 29.9021 16.5 25.0532 18.2114 21.276C20.1338 17.0332 23.5332 13.6338 27.776 11.7114C31.5532 10 36.4021 10 46.1 10C55.7979 10 60.6468 10 64.424 11.7114C68.6669 13.6338 72.0662 17.0332 73.9886 21.276C75.7 25.0532 75.7 29.9021 75.7 39.6C75.7 49.2979 75.7 54.1468 73.9886 57.924C72.0662 62.1669 68.6669 65.5662 64.424 67.4886C60.6468 69.2 55.7979 69.2 46.1 69.2C36.4021 69.2 31.5532 69.2 27.776 67.4886C23.5332 65.5662 20.1338 62.1669 18.2114 57.924C16.5 54.1468 16.5 49.2979 16.5 39.6Z"
                fill="white"
            />
            <path
                d="M16.5 39.6C16.5 29.9021 16.5 25.0532 18.2114 21.276C20.1338 17.0332 23.5332 13.6338 27.776 11.7114C31.5532 10 36.4021 10 46.1 10C55.7979 10 60.6468 10 64.424 11.7114C68.6669 13.6338 72.0662 17.0332 73.9886 21.276C75.7 25.0532 75.7 29.9021 75.7 39.6C75.7 49.2979 75.7 54.1468 73.9886 57.924C72.0662 62.1669 68.6669 65.5662 64.424 67.4886C60.6468 69.2 55.7979 69.2 46.1 69.2C36.4021 69.2 31.5532 69.2 27.776 67.4886C23.5332 65.5662 20.1338 62.1669 18.2114 57.924C16.5 54.1468 16.5 49.2979 16.5 39.6Z"
                fill="url(#paint0_radial_3496_6128)"
                fillOpacity="0.2"
            />
            <path
                d="M17.3 39.6C17.3 34.7392 17.3005 31.137 17.5118 28.2767C17.7224 25.4244 18.1394 23.3733 18.9401 21.6062C20.7824 17.5401 24.0401 14.2824 28.1062 12.4401C29.8733 11.6394 31.9244 11.2224 34.7767 11.0118C37.637 10.8005 41.2392 10.8 46.1 10.8C50.9608 10.8 54.563 10.8005 57.4233 11.0118C60.2756 11.2224 62.3267 11.6394 64.0938 12.4401C68.1599 14.2824 71.4176 17.5401 73.2599 21.6062C74.0606 23.3733 74.4776 25.4244 74.6882 28.2767C74.8995 31.137 74.9 34.7392 74.9 39.6C74.9 44.4608 74.8995 48.063 74.6882 50.9233C74.4776 53.7756 74.0606 55.8267 73.2599 57.5938C71.4176 61.6599 68.1599 64.9176 64.0938 66.7599C62.3267 67.5606 60.2756 67.9776 57.4233 68.1882C54.563 68.3995 50.9608 68.4 46.1 68.4C41.2392 68.4 37.637 68.3995 34.7767 68.1882C31.9244 67.9776 29.8733 67.5606 28.1062 66.7599C24.0401 64.9176 20.7824 61.6599 18.9401 57.5938C18.1394 55.8267 17.7224 53.7756 17.5118 50.9233C17.3005 48.063 17.3 44.4608 17.3 39.6Z"
                stroke="black"
                strokeOpacity="0.05"
                strokeWidth="1.6"
            />
        </g>
        <path
            fillRule="evenodd"
            clipRule="evenodd"
            d="M52.6019 26.8094C54.0687 26.8094 55.2578 28.0165 55.2578 29.5054L55.2578 35.6675C55.2578 37.1564 54.0687 38.3634 52.6019 38.3634C51.1351 38.3634 49.946 37.1564 49.946 35.6675L49.946 29.5054C49.946 28.0165 51.1351 26.8094 52.6019 26.8094Z"
            fill="url(#paint1_angular_3496_6128)"
        />
        <path
            fillRule="evenodd"
            clipRule="evenodd"
            d="M32.8701 33.7408C32.8701 32.2519 34.0592 31.0449 35.526 31.0449H41.5967C43.0635 31.0449 44.2526 32.2519 44.2526 33.7408C44.2526 35.2298 43.0635 36.4368 41.5967 36.4368H35.526C34.0592 36.4368 32.8701 35.2298 32.8701 33.7408Z"
            fill="url(#paint2_angular_3496_6128)"
        />
        <path
            fillRule="evenodd"
            clipRule="evenodd"
            d="M35.9339 43.4703C35.051 42.2919 33.3952 42.0562 32.2266 42.9458C31.0531 43.8392 30.8153 45.529 31.6954 46.7201C31.702 46.729 31.6961 46.7211 31.6969 46.7222L31.6985 46.7243L31.702 46.729L31.7102 46.7399C31.7161 46.7479 31.7231 46.7571 31.7311 46.7676C31.7472 46.7887 31.7673 46.8148 31.7916 46.8457C31.8403 46.9075 31.9055 46.9884 31.9874 47.0857C32.1511 47.2801 32.3822 47.541 32.6805 47.8457C33.2759 48.4536 34.1485 49.2455 35.2989 50.0333C37.6079 51.6147 41.0566 53.1903 45.5821 53.1903C50.1075 53.1903 53.5562 51.6147 55.8653 50.0333C57.0156 49.2455 57.8883 48.4536 58.4836 47.8457C58.782 47.541 59.013 47.2801 59.1767 47.0857C59.2586 46.9884 59.3239 46.9075 59.3725 46.8457C59.3968 46.8148 59.4496 46.7454 59.4656 46.7243C59.4736 46.7138 59.4687 46.7201 59.454 46.7399L59.4621 46.729L59.4656 46.7243L59.4672 46.7222C59.468 46.7211 59.4621 46.729 59.4656 46.7243L59.4687 46.7201C60.3488 45.529 60.111 43.8392 58.9376 42.9458C57.769 42.0562 56.1132 42.2919 55.2303 43.4702C55.2288 43.4722 55.2261 43.4756 55.2222 43.4805C55.2087 43.4977 55.1808 43.5326 55.1386 43.5828C55.054 43.6833 54.913 43.8437 54.7164 44.0444C54.3217 44.4475 53.7122 45.0035 52.8944 45.5637C51.2669 46.6782 48.8347 47.7985 45.5821 47.7985C42.3294 47.7985 39.8972 46.6782 38.2698 45.5637C37.4519 45.0035 36.8425 44.4475 36.4478 44.0444C36.2511 43.8437 36.1102 43.6833 36.0256 43.5828C35.9834 43.5326 35.9554 43.4977 35.9419 43.4805C35.9381 43.4757 35.9354 43.4722 35.9339 43.4703ZM55.2224 43.4808L55.222 43.4813L55.2207 43.4831C55.22 43.484 55.2193 43.485 55.2193 43.485C55.2203 43.4836 55.2213 43.4822 55.2224 43.4808Z"
            fill="url(#paint3_angular_3496_6128)"
        />
        <defs>
            <filter
                id="filter0_dd_3496_6128"
                x="0.5"
                y="0.4"
                width="91.2002"
                height="91.2"
                filterUnits="userSpaceOnUse"
                colorInterpolationFilters="sRGB"
            >
                <feFlood floodOpacity="0" result="BackgroundImageFix" />
                <feColorMatrix
                    in="SourceAlpha"
                    type="matrix"
                    values="0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 127 0"
                    result="hardAlpha"
                />
                <feOffset dy="6.4" />
                <feGaussianBlur stdDeviation="8" />
                <feComposite in2="hardAlpha" operator="out" />
                <feColorMatrix type="matrix" values="0 0 0 0 0.891257 0 0 0 0 0.907635 0 0 0 0 0.956771 0 0 0 1 0" />
                <feBlend mode="normal" in2="BackgroundImageFix" result="effect1_dropShadow_3496_6128" />
                <feColorMatrix
                    in="SourceAlpha"
                    type="matrix"
                    values="0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 127 0"
                    result="hardAlpha"
                />
                <feOffset dy="3.2" />
                <feGaussianBlur stdDeviation="1.6" />
                <feComposite in2="hardAlpha" operator="out" />
                <feColorMatrix type="matrix" values="0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0.05 0" />
                <feBlend mode="normal" in2="effect1_dropShadow_3496_6128" result="effect2_dropShadow_3496_6128" />
                <feBlend mode="normal" in="SourceGraphic" in2="effect2_dropShadow_3496_6128" result="shape" />
            </filter>
            <radialGradient
                id="paint0_radial_3496_6128"
                cx="0"
                cy="0"
                r="1"
                gradientUnits="userSpaceOnUse"
                gradientTransform="translate(39.611 -33.8942) rotate(77.074) scale(57.4399)"
            >
                <stop stopColor="#00DBFF" />
                <stop offset="1" stopColor="#00DBFF" stopOpacity="0" />
            </radialGradient>
            <radialGradient
                id="paint1_angular_3496_6128"
                cx="0"
                cy="0"
                r="1"
                gradientUnits="userSpaceOnUse"
                gradientTransform="translate(45.7186 38.3634) rotate(-9.82972) scale(13.8747 16.9812)"
            >
                <stop offset="0.0576364" stopColor="#FF291F" />
                <stop offset="0.308435" stopColor="#00CBEC" />
                <stop offset="0.642062" stopColor="#A112FF" />
                <stop offset="0.744128" stopColor="#7048E8" />
                <stop offset="0.876835" stopColor="#FF5543" />
            </radialGradient>
            <radialGradient
                id="paint2_angular_3496_6128"
                cx="0"
                cy="0"
                r="1"
                gradientUnits="userSpaceOnUse"
                gradientTransform="translate(46.5013 38.7112) rotate(-168.642) scale(11.5495 11.6547)"
            >
                <stop offset="0.00948703" stopColor="#A112FF" />
                <stop offset="0.308138" stopColor="#FA524E" />
                <stop offset="0.961681" stopColor="#7048E8" />
            </radialGradient>
            <radialGradient
                id="paint3_angular_3496_6128"
                cx="0"
                cy="0"
                r="1"
                gradientUnits="userSpaceOnUse"
                gradientTransform="translate(45.5821 37.1419) rotate(-0.445512) scale(14.4184 13.3628)"
            >
                <stop offset="0.00481802" stopColor="#7048E8" />
                <stop offset="0.417018" stopColor="#00CBEC" />
                <stop offset="0.642062" stopColor="#A112FF" />
                <stop offset="1" stopColor="#FF5543" />
            </radialGradient>
        </defs>
    </svg>
)

export const CodySubscriptionManagePage = withAuthenticatedUser(AuthenticatedCodySubscriptionManagePage)
