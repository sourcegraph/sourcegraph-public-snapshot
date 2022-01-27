import { VisuallyHidden } from '@reach/visually-hidden'
import classNames from 'classnames'
import CloseIcon from 'mdi-react/CloseIcon'
import React from 'react'
import { Button, Input, Modal } from '@sourcegraph/wildcard'
import styles from './InviteMemberModal.module.scss'

export interface InviteMemberModalProps {
    onClose: () => void
    orgName: string
}

export const InviteMemberModal: React.FunctionComponent<InviteMemberModalProps> = props => {
    const { onClose, orgName } = props
    const [member, setMember] = React.useState()
    const title = `Invite teammate to ${orgName}`

    // const handleDeleteSuccess = (): void => {
    //     if (!dashboard.owner) {
    //         history.push('/insights/dashboards')
    //         onClose()
    //         return
    //     }

    //     history.push(`/insights/dashboards/${dashboard.owner.id}`)
    //     onClose()
    // }

    // const { loadingOrError, handler } = useDeleteDashboardHandler({
    //     dashboard,
    //     onSuccess: handleDeleteSuccess,
    // })

    // const isDeleting = !isErrorLike(loadingOrError) && loadingOrError

    return (
        <Modal className={styles.modal} onDismiss={onClose} position="center" aria-label={title}>
            <Button className={classNames('btn-icon', styles.closeButton)} onClick={onClose}>
                <VisuallyHidden>Close</VisuallyHidden>
                <CloseIcon />
            </Button>

            <h2>{title}</h2>

            <Input
                autoFocus
                value={member}
                label="Email address or username"
                title="Email address or username"
                onChange={() => null}
                placeholder="Email address or username"
            />

            <div className="d-flex justify-content-end mt-4">
                <Button type="button" className="mr-2" variant="primary" onClick={onClose}>
                    Send invite
                </Button>
            </div>
        </Modal>
    )
}
