import { mdiCalendarMonth, mdiClose, mdiCreditCardOff, mdiTag, mdiTrendingUp } from '@mdi/js'
import classNames from 'classnames'

import { useMutation } from '@sourcegraph/http-client'
import { Button, H1, H2, Icon, Modal, Text } from '@sourcegraph/wildcard'

import type { AuthenticatedUser } from '../../auth'
import type { ChangeCodyPlanResult, ChangeCodyPlanVariables } from '../../graphql-operations'
import { eventLogger } from '../../tracking/eventLogger'
import { EventName } from '../../util/constants'
import { CodyColorIcon } from '../chat/CodyPageIcon'

import { CHANGE_CODY_PLAN } from './queries'

import styles from './CodySubscriptionPage.module.scss'

export function UpgradeToProModal({
    authenticatedUser,
    onClose,
    onSuccess,
}: {
    authenticatedUser: AuthenticatedUser
    onClose: () => void
    onSuccess: () => void
}): JSX.Element {
    const [changeCodyPlan, { data }] = useMutation<ChangeCodyPlanResult, ChangeCodyPlanVariables>(CHANGE_CODY_PLAN)

    return (
        <Modal isOpen={true} aria-label="Update to Cody Pro" className={styles.upgradeModal} position="center">
            {data?.changeCodyPlan?.codyProEnabled ? (
                <div className="d-flex flex-column justify-content-between align-items-center mby-4 py-4">
                    <CodyColorIcon width={40} height={40} className="mb-4" />
                    <H2>Upgraded to Cody Pro 🎉</H2>
                    <Text>You now have unlimited autocompletions, chat messages and commands.</Text>

                    <Button className="mt-4" variant="primary" onClick={onSuccess}>
                        Get started
                    </Button>
                </div>
            ) : (
                <>
                    <div className="d-flex justify-content-between align-items-center mb-3 border-bottom pb-3">
                        <H2 className="mb-0">Subscription summary</H2>
                        <Icon svgPath={mdiClose} aria-hidden={true} className="cursor-pointer" onClick={onClose} />
                    </div>

                    <div className={classNames('p-3 d-flex', styles.subscriptionSummaryContainer)}>
                        <div className="flex-1 pr-4">
                            <div className="mr-4 border p-3">
                                <div className="border-bottom pb-2 mb-4">
                                    <H1 className={classNames('mb-1', styles.proTitle)}>Pro</H1>
                                    <Text className={classNames('mb-1', styles.proDescription)} size="small">
                                        Best for professional developers
                                    </Text>
                                </div>
                                <div className="mb-1">
                                    <H2 className={classNames('text-muted d-inline mb-0', styles.proPricing)}>$9</H2>
                                    <Text className="mb-0 text-muted d-inline">/month</Text>
                                </div>
                                <Text className="mb-4 text-muted" size="small">
                                    Free until Feb 2024, <strong>no credit card needed</strong>
                                </Text>
                                <Text className="mb-2">
                                    <strong>Unlimited</strong> autocompletes
                                </Text>
                                <Text className="mb-2">
                                    <strong>Unlimited</strong> messages and commands
                                </Text>
                                <Text className="mb-2">Personalization for larger codebases</Text>
                                <Text className="mb-2">Multiple LLM choices for chat</Text>
                            </div>
                        </div>
                        <div className="flex-1">
                            <H2 className="mb-4">About your trial</H2>
                            <div className="d-flex align-items-center mb-3">
                                <div>
                                    <Icon
                                        svgPath={mdiTrendingUp}
                                        className="mr-3 text-primary d-block"
                                        aria-hidden={true}
                                        size="md"
                                    />
                                </div>
                                <div>
                                    <Text weight="bold" className="mb-0">
                                        All limits lifted:
                                    </Text>
                                    <Text className="mb-0" size="small">
                                        Enjoy unrestricted access right away.
                                    </Text>
                                </div>
                            </div>
                            <div className="d-flex align-items-center mb-3">
                                <div>
                                    <Icon
                                        svgPath={mdiCalendarMonth}
                                        className="mr-3 text-primary d-block"
                                        aria-hidden={true}
                                        size="md"
                                    />
                                </div>
                                <div>
                                    <Text weight="bold" className="mb-0">
                                        Trial duration:
                                    </Text>
                                    <Text className="mb-0" size="small">
                                        Your trial runs until <strong>February 14, 2024.</strong>
                                    </Text>
                                </div>
                            </div>
                            <div className="d-flex align-items-center mb-3">
                                <div>
                                    <Icon
                                        svgPath={mdiCreditCardOff}
                                        className="mr-3 text-primary d-block"
                                        aria-hidden={true}
                                        size="md"
                                    />
                                </div>
                                <div>
                                    <Text weight="bold" className="mb-0">
                                        No credit card required:
                                    </Text>
                                    <Text className="mb-0" size="small">
                                        We'll reach out for billing details before your trial ends.
                                    </Text>
                                </div>
                            </div>
                            <div className="d-flex align-items-center mb-4">
                                <div>
                                    <Icon
                                        svgPath={mdiTag}
                                        className="mr-3 text-primary d-block"
                                        aria-hidden={true}
                                        size="md"
                                    />
                                </div>
                                <div>
                                    <Text weight="bold" className="mb-0">
                                        No hidden fees, no surprises
                                    </Text>
                                    <Text className="mb-0" size="small">
                                        We're eager to have you onboard and listen to your feedback
                                    </Text>
                                </div>
                            </div>
                            <div className="d-flex justify-content-center mt-4 pt-4">
                                <Button
                                    variant="primary"
                                    onClick={() => {
                                        eventLogger.log(EventName.CODY_SUBSCRIPTION_PLAN_CONFIRMED, {
                                            tier: 'pro',
                                        })

                                        changeCodyPlan({ variables: { pro: true, id: authenticatedUser.id } })
                                    }}
                                >
                                    <Icon svgPath={mdiTrendingUp} className="mr-1" aria-hidden={true} />
                                    Start trial
                                </Button>
                            </div>
                        </div>
                    </div>
                </>
            )}
        </Modal>
    )
}
