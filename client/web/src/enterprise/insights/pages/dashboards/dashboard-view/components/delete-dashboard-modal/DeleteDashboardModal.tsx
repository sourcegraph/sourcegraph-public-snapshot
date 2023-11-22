import React from 'react'

import { mdiClose } from '@mdi/js'
import { VisuallyHidden } from '@reach/visually-hidden'
import { useNavigate } from 'react-router-dom'

import { isErrorLike } from '@sourcegraph/common'
import { Button, Modal, H2, Icon, ErrorAlert } from '@sourcegraph/wildcard'

import { LoaderButton } from '../../../../../../../components/LoaderButton'
import type { CustomInsightDashboard } from '../../../../../core/types'

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
    const navigate = useNavigate()

    const handleDeleteSuccess = (): void => {
        navigate('/insights/dashboards')
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
                <Icon svgPath={mdiClose} inline={false} aria-hidden={true} />
            </Button>

            <H2 className="text-danger">Delete ”{dashboard.title}”</H2>

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
