import React, { useContext, useMemo } from 'react'

import { VisuallyHidden } from '@reach/visually-hidden'
import CloseIcon from 'mdi-react/CloseIcon'

import { asError } from '@sourcegraph/common'
import { Button, LoadingSpinner, useObservable, Modal, Typography } from '@sourcegraph/wildcard'

import { FORM_ERROR, SubmissionErrors } from '../../../../../components/form/hooks/useForm'
import { CodeInsightsBackendContext, CustomInsightDashboard } from '../../../../../core'

import {
    AddInsightFormValues,
    AddInsightModalContent,
} from './components/add-insight-modal-content/AddInsightModalContent'

import styles from './AddInsightModal.module.scss'

export interface AddInsightModalProps {
    dashboard: CustomInsightDashboard
    onClose: () => void
}

export const AddInsightModal: React.FunctionComponent<React.PropsWithChildren<AddInsightModalProps>> = props => {
    const { dashboard, onClose } = props
    const { getAccessibleInsightsList, assignInsightsToDashboard } = useContext(CodeInsightsBackendContext)

    const insights = useObservable(useMemo(() => getAccessibleInsightsList(), [getAccessibleInsightsList]))

    const initialValues = useMemo<AddInsightFormValues>(
        () => ({
            searchInput: '',
            insightIds: dashboard.insightIds ?? [],
        }),
        [dashboard]
    )

    const handleSubmit = async (values: AddInsightFormValues): Promise<void | SubmissionErrors> => {
        try {
            const { insightIds } = values

            await assignInsightsToDashboard({
                id: dashboard.id,
                prevInsightIds: dashboard.insightIds ?? [],
                nextInsightIds: insightIds,
            }).toPromise()

            onClose()
        } catch (error) {
            return { [FORM_ERROR]: asError(error) }
        }
    }

    if (insights === undefined) {
        return (
            <Modal className={styles.modal} aria-label="Add insights to dashboard modal">
                <LoadingSpinner inline={false} />
            </Modal>
        )
    }

    return (
        <Modal className={styles.modal} onDismiss={onClose} aria-label="Add insights to dashboard modal">
            <Button variant="icon" className={styles.closeButton} onClick={onClose}>
                <VisuallyHidden>Close</VisuallyHidden>
                <CloseIcon />
            </Button>

            <Typography.H2 className="mb-3">
                Add insight to <q>{dashboard.title}</q>
            </Typography.H2>

            {!insights.length && <span>There are no insights for this dashboard.</span>}

            {insights.length > 0 && (
                <AddInsightModalContent
                    initialValues={initialValues}
                    insights={insights}
                    dashboardID={dashboard.id}
                    onCancel={onClose}
                    onSubmit={handleSubmit}
                />
            )}
        </Modal>
    )
}
