import React, { useEffect, useState } from 'react'
import type { ReactElement } from 'react'

import { mdiTrendingUp } from '@mdi/js'
import classNames from 'classnames'

import { useQuery } from '@sourcegraph/http-client'
import { Icon, PageHeader, Button, H1, H2, H3, Text, ButtonLink, useSearchParameters, H4 } from '@sourcegraph/wildcard'

import type { AuthenticatedUser } from '../../auth'
import { Page } from '../../components/Page'
import { PageTitle } from '../../components/PageTitle'
import { useFeatureFlag } from '../../featureFlags/useFeatureFlag'
import type { UserCodyPlanResult, UserCodyPlanVariables } from '../../graphql-operations'
import { eventLogger } from '../../tracking/eventLogger'
import { EventName } from '../../util/constants'
import { CodyColorIcon } from '../chat/CodyPageIcon'
import { isCodyEnabled } from '../isCodyEnabled'

import { CancelProModal } from './CancelProModal'
import { USER_CODY_PLAN } from './queries'
import { UpgradeToProModal } from './UpgradeToProModal'

import styles from './CodySubscriptionPage.module.scss'

interface CodySubscriptionPageProps {
    isSourcegraphDotCom: boolean
    authenticatedUser?: AuthenticatedUser | null
}

export const CodySubscriptionPage: React.FunctionComponent<CodySubscriptionPageProps> = ({
    isSourcegraphDotCom,
    authenticatedUser,
}) => {
    const parameters = useSearchParameters()

    const utm_source = parameters.get('utm_source')

    useEffect(() => {
        eventLogger.log(EventName.CODY_SUBSCRIPTION_PAGE_VIEWED, { utm_source })
    }, [utm_source])

    const { data } = useQuery<UserCodyPlanResult, UserCodyPlanVariables>(USER_CODY_PLAN, {})

    const [isEnabled] = useFeatureFlag('cody-pro', false)
    const [showUpgradeToPro, setShowUpgradeToPro] = useState<boolean>(false)
    const [showCancelPro, setShowCancelPro] = useState<boolean>(false)

    if (!isCodyEnabled() || !isSourcegraphDotCom || !isEnabled || !data?.currentUser || !authenticatedUser) {
        return null
    }

    const { codyProEnabled } = data.currentUser

    return (
        <>
            <Page className={classNames('d-flex flex-column')}>
                <PageTitle title="Cody Subscription" />
                <PageHeader
                    className="mb-4"
                    actions={
                        <ButtonLink to="/cody/manage" variant="secondary" outline={true} size="sm">
                            Dashboard
                        </ButtonLink>
                    }
                >
                    <PageHeader.Heading as="h2" styleAs="h1">
                        <div className="d-inline-flex align-items-center">
                            <CodyColorIcon width={40} height={40} className="mr-2" /> Subscription Plans
                        </div>
                    </PageHeader.Heading>
                </PageHeader>

                <div className={classNames('d-flex mt-4', styles.responsiveContainer)}>
                    <div className="border d-flex flex-column flex-1 bg-1">
                        {!codyProEnabled && (
                            <div className="bg-1 border-bottom d-flex justify-content-center align-items-center bg-2">
                                <H4 className="mb-2 mt-2">Your Plan</H4>
                            </div>
                        )}
                        <div className="p-3">
                            <div className="border-bottom pb-3">
                                <H2 className="mb-1 text-muted">Community</H2>
                                <Text className="mb-0 text-muted" size="small">
                                    Best for hobbyists
                                </Text>
                            </div>
                            <div className="border-bottom py-4">
                                <H1 className="mb-1 text-muted">Free</H1>
                            </div>
                            <div className="border-bottom py-4">
                                <Text weight="bold" className="d-inline">
                                    500
                                </Text>{' '}
                                <Text className="d-inline text-muted">Autocompletions per month</Text>
                            </div>
                            <div className="border-bottom py-4">
                                <Text weight="bold" className="d-inline">
                                    20
                                </Text>{' '}
                                <Text className="d-inline text-muted">Messages and Commands per month</Text>
                            </div>
                            <div className="border-bottom py-4">
                                <div className="mb-1">
                                    <Text weight="bold" className="d-inline mb-0">
                                        Limited
                                    </Text>{' '}
                                    <Text className="d-inline text-muted mb-0">Private Code Embeddings</Text>
                                </div>
                                <Text className="mb-1 text-muted">Current project with keyword search</Text>
                                <Text className="mb-1 text-muted">Embeddings on some public repo</Text>
                            </div>
                            <div className="border-bottom py-4">
                                <Text weight="bold" className="mb-1">
                                    Code editor support
                                </Text>{' '}
                                <Text className="d-inline text-muted">All supported Code Editors</Text>
                            </div>
                            <div className="border-bottom py-4">
                                <Text weight="bold" className="mb-1">
                                    Support
                                </Text>{' '}
                                <Text className="d-inline text-muted">Community support only</Text>
                            </div>
                        </div>
                    </div>
                    <div className={classNames('border d-flex flex-column flex-1 bg-1', styles.proContainer)}>
                        {codyProEnabled && (
                            <div className="bg-1 border-bottom d-flex justify-content-center align-items-center bg-2">
                                <H4 className="mb-2 mt-2">Your Plan</H4>
                            </div>
                        )}
                        <div className="p-3">
                            <div className="border-bottom pb-2">
                                <H1 className={classNames('mb-1', styles.proTitle)}>Pro</H1>
                                <Text className={classNames('mb-1 text-primary', styles.proDescription)} size="base">
                                    Best for professional developers
                                </Text>
                            </div>
                            <div className="d-flex flex-column border-bottom py-4">
                                <div className="mb-1">
                                    <H2 className={classNames('text-muted d-inline mb-0', styles.proPricing)}>22$</H2>
                                    <Text className="mb-0 text-muted d-inline">/ month</Text>
                                </div>
                                <Text className="mb-3 text-muted" size="small">
                                    Free until Feb 2024, <strong>no credit card needed</strong>
                                </Text>
                                {codyProEnabled ? (
                                    <div>
                                        <Text
                                            className="mb-0 text-muted d-inline cursor-pointer"
                                            size="small"
                                            onClick={() => setShowCancelPro(true)}
                                        >
                                            Cancel
                                        </Text>
                                    </div>
                                ) : (
                                    <Button
                                        className="flex-1"
                                        variant="primary"
                                        onClick={() => setShowUpgradeToPro(true)}
                                    >
                                        <Icon svgPath={mdiTrendingUp} className="mr-1" aria-hidden={true} />
                                        Get Pro Trial
                                    </Button>
                                )}
                            </div>
                            <div className="border-bottom py-4">
                                <Text weight="bold" className={classNames('d-inline', styles.amazing)}>
                                    Unlimited
                                </Text>{' '}
                                <Text className="d-inline text-muted">Autocompletions per month</Text>
                            </div>
                            <div className="border-bottom py-4">
                                <Text weight="bold" className={classNames('d-inline', styles.amazing)}>
                                    Unlimited
                                </Text>{' '}
                                <Text className="d-inline text-muted">Messages and Commands per month</Text>
                            </div>
                            <div className="border-bottom py-4">
                                <div className="mb-1">
                                    <Text weight="bold" className={classNames('d-inline mb-0', styles.amazing)}>
                                        Unlimited
                                    </Text>{' '}
                                    <Text className="d-inline text-muted mb-0">Private Code Embeddings</Text>
                                </div>
                                <Text className="mb-1 text-muted">Current project with keyword search</Text>
                                <Text className="mb-0 text-muted">Embeddings on some public repo</Text>
                            </div>
                            <div className="border-bottom py-4">
                                <Text weight="bold" className="mb-1">
                                    Code editor support
                                </Text>{' '}
                                <Text className="d-inline text-muted">All supported Code Editors</Text>
                            </div>
                            <div className="border-bottom py-4">
                                <Text weight="bold" className="mb-1">
                                    Support
                                </Text>{' '}
                                <Text className="d-inline text-muted">Community support only</Text>
                            </div>
                        </div>
                    </div>
                    <div className="border d-flex flex-column flex-1 bg-1 border p-3">
                        <div className="border-bottom pb-3">
                            <H1 className={classNames('mb-1', styles.enterpriseColor)}>Enterprise</H1>
                            <Text className={classNames('mb-0', styles.enterpriseColorLight)} size="small">
                                Best for teams
                            </Text>
                        </div>
                        <div className="d-flex flex-column border-bottom py-4">
                            <H3 className={classNames('text-muted', codyProEnabled ? 'mb-2' : 'mb-4')}>
                                Custom pricing
                            </H3>
                            <ButtonLink className="flex-1 mt-3" variant="secondary" outline={true}>
                                Contact Sales
                            </ButtonLink>
                        </div>
                        <div className="border-bottom py-4">
                            <Text weight="bold" className={classNames('d-inline', styles.amazing)}>
                                Unlimited
                            </Text>{' '}
                            <Text className="d-inline text-muted">Autocompletions per month</Text>
                        </div>
                        <div className="border-bottom py-4">
                            <Text weight="bold" className={classNames('d-inline', styles.amazing)}>
                                Unlimited
                            </Text>{' '}
                            <Text className="d-inline text-muted">Messages and Commands per month</Text>
                        </div>
                        <div className="border-bottom py-4">
                            <div className="mb-1">
                                <Text weight="bold" className={classNames('d-inline mb-0', styles.amazing)}>
                                    Unlimited
                                </Text>{' '}
                                <Text className="d-inline text-muted mb-0">Private Code Embeddings</Text>
                            </div>
                            <Text className="mb-1 text-muted">Current project with keyword search</Text>
                            <Text className="mb-0 text-muted">Embeddings on some public repo</Text>
                        </div>
                        <div className="border-bottom py-4">
                            <Text weight="bold" className="mb-1">
                                Enterprise Features
                            </Text>
                            <Text className="mb-1 text-muted">Bring your own LLM</Text>
                            <Text className="mb-0 text-muted">Single Tenant</Text>
                            <Text className="mb-0 text-muted">SAML / SSO</Text>
                            <Text className="mb-0 text-muted">Guardrails</Text>
                        </div>
                        <div className="border-bottom py-4">
                            <Text weight="bold" className="mb-1">
                                Code editor support
                            </Text>{' '}
                            <Text className="d-inline text-muted">All supported Code Editors</Text>
                        </div>
                        <div className="py-4">
                            <Text weight="bold" className="mb-1">
                                Support
                            </Text>{' '}
                            <Text className="d-inline text-muted">Community support only</Text>
                        </div>
                    </div>
                </div>
            </Page>
            {showUpgradeToPro && (
                <UpgradeToProModal
                    onClose={() => {
                        setShowUpgradeToPro(false)
                    }}
                    authenticatedUser={authenticatedUser}
                />
            )}
            {showCancelPro && (
                <CancelProModal
                    onClose={() => {
                        setShowCancelPro(false)
                    }}
                    authenticatedUser={authenticatedUser}
                />
            )}
        </>
    )
}

export const ProTierIcon = ({ className }: { className?: string }): ReactElement => (
    <svg width="92" height="92" viewBox="0 0 92 92" fill="none" className={className}>
        <circle cx="46" cy="46" r="46" fill="url(#paint0_radial_2897_1551)" />
        <g filter="url(#filter0_d_2897_1551)">
            <path
                d="M19.902 56V28.3636H30.8054C32.9015 28.3636 34.6873 28.764 36.1626 29.5646C37.638 30.3563 38.7625 31.4583 39.5362 32.8707C40.3189 34.2741 40.7102 35.8935 40.7102 37.7287C40.7102 39.5639 40.3144 41.1832 39.5227 42.5866C38.7311 43.9901 37.584 45.0831 36.0817 45.8658C34.5883 46.6484 32.7801 47.0398 30.657 47.0398H23.7074V42.3572H29.7124C30.8369 42.3572 31.7635 42.1638 32.4922 41.777C33.2299 41.3812 33.7786 40.8369 34.1385 40.1442C34.5073 39.4425 34.6918 38.6373 34.6918 37.7287C34.6918 36.8111 34.5073 36.0104 34.1385 35.3267C33.7786 34.634 33.2299 34.0987 32.4922 33.7209C31.7545 33.334 30.8189 33.1406 29.6854 33.1406H25.745V56H19.902ZM41.3975 56V35.2727H46.9707V38.8892H47.1866C47.5645 37.6027 48.1987 36.6312 49.0893 35.9744C49.9799 35.3087 51.0055 34.9759 52.166 34.9759C52.4539 34.9759 52.7643 34.9938 53.0971 35.0298C53.43 35.0658 53.7224 35.1153 53.9743 35.1783V40.2791C53.7044 40.1982 53.331 40.1262 52.8542 40.0632C52.3774 40.0002 51.9411 39.9688 51.5453 39.9688C50.6996 39.9688 49.944 40.1532 49.2782 40.522C48.6215 40.8819 48.0997 41.3857 47.7129 42.0334C47.335 42.6811 47.1461 43.4278 47.1461 44.2734V56H41.3975ZM62.3466 56.4048C60.2505 56.4048 58.4377 55.9595 56.9084 55.0689C55.388 54.1693 54.214 52.9188 53.3864 51.3175C52.5587 49.7071 52.1449 47.8404 52.1449 45.7173C52.1449 43.5762 52.5587 41.705 53.3864 40.1037C54.214 38.4934 55.388 37.2429 56.9084 36.3523C58.4377 35.4527 60.2505 35.0028 62.3466 35.0028C64.4427 35.0028 66.2509 35.4527 67.7713 36.3523C69.3007 37.2429 70.4792 38.4934 71.3068 40.1037C72.1345 41.705 72.5483 43.5762 72.5483 45.7173C72.5483 47.8404 72.1345 49.7071 71.3068 51.3175C70.4792 52.9188 69.3007 54.1693 67.7713 55.0689C66.2509 55.9595 64.4427 56.4048 62.3466 56.4048ZM62.3736 51.9517C63.3272 51.9517 64.1233 51.6818 64.7621 51.142C65.4008 50.5933 65.8821 49.8466 66.206 48.902C66.5388 47.9574 66.7053 46.8823 66.7053 45.6768C66.7053 44.4714 66.5388 43.3963 66.206 42.4517C65.8821 41.5071 65.4008 40.7604 64.7621 40.2116C64.1233 39.6629 63.3272 39.3885 62.3736 39.3885C61.411 39.3885 60.6013 39.6629 59.9446 40.2116C59.2969 40.7604 58.8066 41.5071 58.4737 42.4517C58.1499 43.3963 57.9879 44.4714 57.9879 45.6768C57.9879 46.8823 58.1499 47.9574 58.4737 48.902C58.8066 49.8466 59.2969 50.5933 59.9446 51.142C60.6013 51.6818 61.411 51.9517 62.3736 51.9517Z"
                fill="#EFF2F5"
            />
            <path
                d="M19.902 56V28.3636H30.8054C32.9015 28.3636 34.6873 28.764 36.1626 29.5646C37.638 30.3563 38.7625 31.4583 39.5362 32.8707C40.3189 34.2741 40.7102 35.8935 40.7102 37.7287C40.7102 39.5639 40.3144 41.1832 39.5227 42.5866C38.7311 43.9901 37.584 45.0831 36.0817 45.8658C34.5883 46.6484 32.7801 47.0398 30.657 47.0398H23.7074V42.3572H29.7124C30.8369 42.3572 31.7635 42.1638 32.4922 41.777C33.2299 41.3812 33.7786 40.8369 34.1385 40.1442C34.5073 39.4425 34.6918 38.6373 34.6918 37.7287C34.6918 36.8111 34.5073 36.0104 34.1385 35.3267C33.7786 34.634 33.2299 34.0987 32.4922 33.7209C31.7545 33.334 30.8189 33.1406 29.6854 33.1406H25.745V56H19.902ZM41.3975 56V35.2727H46.9707V38.8892H47.1866C47.5645 37.6027 48.1987 36.6312 49.0893 35.9744C49.9799 35.3087 51.0055 34.9759 52.166 34.9759C52.4539 34.9759 52.7643 34.9938 53.0971 35.0298C53.43 35.0658 53.7224 35.1153 53.9743 35.1783V40.2791C53.7044 40.1982 53.331 40.1262 52.8542 40.0632C52.3774 40.0002 51.9411 39.9688 51.5453 39.9688C50.6996 39.9688 49.944 40.1532 49.2782 40.522C48.6215 40.8819 48.0997 41.3857 47.7129 42.0334C47.335 42.6811 47.1461 43.4278 47.1461 44.2734V56H41.3975ZM62.3466 56.4048C60.2505 56.4048 58.4377 55.9595 56.9084 55.0689C55.388 54.1693 54.214 52.9188 53.3864 51.3175C52.5587 49.7071 52.1449 47.8404 52.1449 45.7173C52.1449 43.5762 52.5587 41.705 53.3864 40.1037C54.214 38.4934 55.388 37.2429 56.9084 36.3523C58.4377 35.4527 60.2505 35.0028 62.3466 35.0028C64.4427 35.0028 66.2509 35.4527 67.7713 36.3523C69.3007 37.2429 70.4792 38.4934 71.3068 40.1037C72.1345 41.705 72.5483 43.5762 72.5483 45.7173C72.5483 47.8404 72.1345 49.7071 71.3068 51.3175C70.4792 52.9188 69.3007 54.1693 67.7713 55.0689C66.2509 55.9595 64.4427 56.4048 62.3466 56.4048ZM62.3736 51.9517C63.3272 51.9517 64.1233 51.6818 64.7621 51.142C65.4008 50.5933 65.8821 49.8466 66.206 48.902C66.5388 47.9574 66.7053 46.8823 66.7053 45.6768C66.7053 44.4714 66.5388 43.3963 66.206 42.4517C65.8821 41.5071 65.4008 40.7604 64.7621 40.2116C64.1233 39.6629 63.3272 39.3885 62.3736 39.3885C61.411 39.3885 60.6013 39.6629 59.9446 40.2116C59.2969 40.7604 58.8066 41.5071 58.4737 42.4517C58.1499 43.3963 57.9879 44.4714 57.9879 45.6768C57.9879 46.8823 58.1499 47.9574 58.4737 48.902C58.8066 49.8466 59.2969 50.5933 59.9446 51.142C60.6013 51.6818 61.411 51.9517 62.3736 51.9517Z"
                fill="url(#paint1_angular_2897_1551)"
            />
        </g>
        <defs>
            <filter
                id="filter0_d_2897_1551"
                x="18.9019"
                y="28.3635"
                width="54.6465"
                height="30.0413"
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
                <feOffset dy="1" />
                <feGaussianBlur stdDeviation="0.5" />
                <feComposite in2="hardAlpha" operator="out" />
                <feColorMatrix type="matrix" values="0 0 0 0 0.278089 0 0 0 0 0.267405 0 0 0 0 0.267405 0 0 0 0.25 0" />
                <feBlend mode="normal" in2="BackgroundImageFix" result="effect1_dropShadow_2897_1551" />
                <feBlend mode="normal" in="SourceGraphic" in2="effect1_dropShadow_2897_1551" result="shape" />
            </filter>
            <radialGradient
                id="paint0_radial_2897_1551"
                cx="0"
                cy="0"
                r="1"
                gradientUnits="userSpaceOnUse"
                gradientTransform="translate(46 46) rotate(66.2277) scale(45.8939)"
            >
                <stop stopColor="#FF0000" stopOpacity="0.1" />
                <stop offset="1" stopColor="#D9D9D9" stopOpacity="0" />
            </radialGradient>
            <radialGradient
                id="paint1_angular_2897_1551"
                cx="0"
                cy="0"
                r="1"
                gradientUnits="userSpaceOnUse"
                gradientTransform="translate(47.2619 46.4966) rotate(-8.02941) scale(54.9809 28.665)"
            >
                <stop stopColor="#EC4D49" />
                <stop offset="0.262672" stopColor="#7048E8" />
                <stop offset="0.465801" stopColor="#4AC1E8" />
                <stop offset="0.752264" stopColor="#4D0B79" />
                <stop offset="1" stopColor="#FF5543" />
            </radialGradient>
        </defs>
    </svg>
)
