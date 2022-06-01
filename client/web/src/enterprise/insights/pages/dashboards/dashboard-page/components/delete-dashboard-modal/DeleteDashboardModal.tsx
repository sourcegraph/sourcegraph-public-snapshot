import React from 'react'

import { VisuallyHidden } from '@reach/visually-hidden'
import CloseIcon from 'mdi-react/CloseIcon'
import { useHistory } from 'react-router-dom'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { isErrorLike } from '@sourcegraph/common'
import { Button, Modal, Typography } from '@sourcegraph/wildcard'

import { LoaderButton } from '../../../../../../../components/LoaderButton'
import { CustomInsightDashboard } from '../../../../../core/types'

import { useDeleteDashboardHandler } from './hooks/use-delete-dashboard-handler'

import styles from './DeleteDashobardModal.module.scss'

export interface DeleteDashboardModalProps {
    dashboard: CustomInsightDashboard
    onClose: () => void
}

export const DeleteDashboardModal: React.FunctionComponent<
    React.PropsWithChildren<DeleteDashboardModalProps>
> = props => {
    const { dashboard, onClose } = props
    const history = useHistory()

    const handleDeleteSuccess = (): void => {
        history.push('/insights/dashboards')
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

            <Typography.H2 className="text-danger">Delete ”{dashboard.title}”</Typography.H2>

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
