import { VisuallyHidden } from '@reach/visually-hidden'
import CloseIcon from 'mdi-react/CloseIcon'
import React from 'react'
import { useHistory } from 'react-router-dom'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { isErrorLike } from '@sourcegraph/common'
import { Button, Modal } from '@sourcegraph/wildcard'

import { LoaderButton } from '../../../../../../../components/LoaderButton'
import { CustomInsightDashboard } from '../../../../../core/types'

import styles from './DeleteDashobardModal.module.scss'
import { useDeleteDashboardHandler } from './hooks/use-delete-dashboard-handler'

export interface DeleteDashboardModalProps {
    dashboard: CustomInsightDashboard
    onClose: () => void
}

export const DeleteDashboardModal: React.FunctionComponent<DeleteDashboardModalProps> = props => {
    const { dashboard, onClose } = props
    const history = useHistory()

    const handleDeleteSuccess = (): void => {
        if (!dashboard.owner) {
            history.push('/insights/dashboards')
            onClose()
            return
        }

        history.push(`/insights/dashboards/${dashboard.owner.id}`)
        onClose()
    }

    const { loadingOrError, handler } = useDeleteDashboardHandler({
        dashboard,
        onSuccess: handleDeleteSuccess,
    })

    const isDeleting = !isErrorLike(loadingOrError) && loadingOrError

    return (
        <Modal className={styles.modal} onDismiss={onClose} aria-label="Delete code insight dashboard modal">
            <Button variant="icon" className={styles.closeButton} onClick={onClose}>
                <VisuallyHidden>Close</VisuallyHidden>
                <CloseIcon />
            </Button>

            <h2 className="text-danger">Delete ”{dashboard.title}”</h2>

            <span className="d-block mb-4">
                This can't be undone. You will still be able to access insights from this dashboard in ”All insights”.
            </span>

            {isErrorLike(loadingOrError) && <ErrorAlert className='className="mt-3"' error={loadingOrError} />}

            <div className="d-flex justify-content-end mt-4">
                <Button type="button" className="mr-2" variant="secondary" onClick={onClose}>
                    Cancel
                </Button>

                <LoaderButton
                    alwaysShowLabel={true}
                    loading={isDeleting}
                    label={isDeleting ? 'Deleting' : 'Delete forever'}
                    disabled={isDeleting}
                    onClick={handler}
                    variant="danger"
                />
            </div>
        </Modal>
    )
}
