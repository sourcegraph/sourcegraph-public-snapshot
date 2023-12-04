import { useMutation } from '@sourcegraph/http-client'
import { Modal, Button, H2, Text } from '@sourcegraph/wildcard'

import type { AuthenticatedUser } from '../../auth'
import type { ChangeCodyPlanResult, ChangeCodyPlanVariables } from '../../graphql-operations'

import { CHANGE_CODY_PLAN } from './queries'

import styles from './CodySubscriptionPage.module.scss'

export function CancelProModal({
    authenticatedUser,
    onClose,
}: {
    authenticatedUser: AuthenticatedUser
    onClose: () => void
}): JSX.Element {
    const [changeCodyPlan, { data }] = useMutation<ChangeCodyPlanResult, ChangeCodyPlanVariables>(CHANGE_CODY_PLAN)

    return (
        <Modal isOpen={true} aria-label="Update to Cody Pro" className={styles.cancelModal} position="center">
            {data && !data.changeCodyPlan?.codyProEnabled ? (
                <div className="d-flex flex-column py-2">
                    <H2>Sorry to see you go.</H2>

                    <Text>
                        Feel free to continue to use Cody Free, and you're always welcome to upgrade to Cody Pro at any
                        time to get unlimited chats, commands, and autocomplete suggestions.
                    </Text>

                    <div>
                        <Button className="mt-2" variant="primary" onClick={onClose}>
                            Close
                        </Button>
                    </div>
                </div>
            ) : (
                <div className="d-flex flex-column py-2">
                    <H2>Are you sure you want to cancel your Cody Pro plan?</H2>

                    <Text>
                        You'll still be able to access Cody Free, but you'll lose access to unlimited chats, commands,
                        and autocomplete suggestions. Your 1 GB embeddings limit will also decrease to 200 MB.
                    </Text>

                    <div className="mt-2">
                        <Button className="mr-2" variant="primary" autoFocus={true} onClick={onClose}>
                            Stay on Cody Pro
                        </Button>
                        <Button
                            variant="secondary"
                            onClick={() => changeCodyPlan({ variables: { pro: false, id: authenticatedUser.id } })}
                        >
                            Switch to Free
                        </Button>
                    </div>
                </div>
            )}
        </Modal>
    )
}
