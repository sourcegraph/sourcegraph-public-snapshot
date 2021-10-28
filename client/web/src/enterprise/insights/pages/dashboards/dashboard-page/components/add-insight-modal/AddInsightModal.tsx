import Dialog from '@reach/dialog'
import { VisuallyHidden } from '@reach/visually-hidden'
import classNames from 'classnames'
import CloseIcon from 'mdi-react/CloseIcon'
import React, { useContext, useMemo } from 'react'
import { of } from 'rxjs'

import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { asError } from '@sourcegraph/shared/src/util/errors'
import { useObservable } from '@sourcegraph/shared/src/util/useObservable'

import { FORM_ERROR, SubmissionErrors } from '../../../../../components/form/hooks/useForm'
import { CodeInsightsBackendContext } from '../../../../../core/backend/code-insights-backend-context'
import { SettingsBasedInsightDashboard } from '../../../../../core/types'

import styles from './AddInsightModal.module.scss'
import {
    AddInsightFormValues,
    AddInsightModalContent,
} from './components/add-insight-modal-content/AddInsightModalContent'

export interface AddInsightModalProps {
    dashboard: SettingsBasedInsightDashboard
    onClose: () => void
}

export const AddInsightModal: React.FunctionComponent<AddInsightModalProps> = props => {
    const { dashboard, onClose } = props
    const { getReachableInsights, updateDashboard } = useContext(CodeInsightsBackendContext)

    if (!dashboard.owner) {
        throw new Error('TODO: Update this to visibility default')
    }

    const insights = useObservable(
        useMemo(() => getReachableInsights(dashboard.owner!.id) || of([]), [dashboard.owner, getReachableInsights])
    )

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

            if (!dashboard.owner) {
                throw new Error('TODO: Update this to visibility default')
            }

            await updateDashboard({
                previousDashboard: dashboard,
                nextDashboardInput: {
                    name: dashboard.title,
                    visibility: dashboard.owner.id,
                    insightIds,
                },
            }).toPromise()
            onClose()
        } catch (error) {
            return { [FORM_ERROR]: asError(error) }
        }
    }

    if (insights === undefined) {
        return (
            <Dialog className={styles.modal} aria-label="Add insights to dashboard modal">
                <LoadingSpinner />
            </Dialog>
        )
    }

    return (
        <Dialog className={styles.modal} onDismiss={onClose} aria-label="Add insights to dashboard modal">
            <button type="button" className={classNames('btn btn-icon', styles.closeButton)} onClick={onClose}>
                <VisuallyHidden>Close</VisuallyHidden>
                <CloseIcon />
            </button>

            <h2 className="mb-3">
                Add insight to <q>{dashboard.title}</q>
            </h2>

            {!insights.length && <span>There are no insights for this dashboard.</span>}

            {insights.length > 0 && (
                <AddInsightModalContent
                    initialValues={initialValues}
                    insights={insights}
                    onCancel={onClose}
                    onSubmit={handleSubmit}
                />
            )}
        </Dialog>
    )
}
