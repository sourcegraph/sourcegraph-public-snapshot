import { VisuallyHidden } from '@reach/visually-hidden'
import classNames from 'classnames'
import CloseIcon from 'mdi-react/CloseIcon'
import React from 'react'
import { useHistory } from 'react-router-dom'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { isErrorLike } from '@sourcegraph/common'
import { Button, Modal } from '@sourcegraph/wildcard'

import { LoaderButton } from '../../components/LoaderButton'

import styles from './InviteMemberModal.module.scss'

export interface InviteMemberModalProps {
    onClose: () => void
    orgName: string
}

export const InviteMemberModal: React.FunctionComponent<InviteMemberModalProps> = props => {
    const { onClose, orgName } = props
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

            <span className="d-block mb-4">
                This can't be undone. You will still be able to access insights from this dashboard in ”All insights”.
            </span>

            <div className="d-flex justify-content-end mt-4">
                <Button type="button" className="mr-2" variant="primary" onClick={onClose}>
                    Send invite
                </Button>
            </div>
        </Modal>
    )
}
