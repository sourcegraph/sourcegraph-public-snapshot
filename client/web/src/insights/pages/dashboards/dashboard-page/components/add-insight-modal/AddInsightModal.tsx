import Dialog from '@reach/dialog';
import { VisuallyHidden } from '@reach/visually-hidden';
import classnames from 'classnames';
import CloseIcon from 'mdi-react/CloseIcon';
import React from 'react'

import { Form } from '@sourcegraph/branded/src/components/Form';
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings';

import { Insight, RealInsightDashboard } from '../../../../../core/types';

import styles from './AddInsightModal.module.scss'
import { useReachableInsights } from './hooks/use-reachable-insights';

export interface AddInsightModalProps extends SettingsCascadeProps {
    dashboard: RealInsightDashboard
    onClose: () => void
}

export const AddInsightModal: React.FunctionComponent<AddInsightModalProps> = props => {
    const { dashboard, settingsCascade, onClose } = props

    const insights = useReachableInsights({ ownerId: dashboard.owner.id, settingsCascade })

    return (
        <Dialog className={styles.modal} onDismiss={close}>
            <button
                type='button'
                className={classnames('btn btn-icon', styles.closeButton)}
                onClick={onClose}>

                <VisuallyHidden>Close</VisuallyHidden>
                <CloseIcon/>
            </button>

            <h2>Add insight to the <code>{ dashboard.title }</code> dashboard</h2>

            <h3>Insights</h3>

            {
                !insights.length &&
                    <span>There are no insights for this dashboard.</span>
            }

            {
                insights.length && <AddInsightModalContent insights={insights}/>
            }
        </Dialog>
    )
}

interface AddInsightModalContentProps {
    insights: Insight[]
}

const AddInsightModalContent: React.FunctionComponent<AddInsightModalContentProps> = props => {
    const { insights } = props

    return (
        <Form>

            { insights.map(insight =>
                <label key={insight.id}>

                    <span>{insight.title}</span>

                    <input
                        type="checkbox"
                        name={insight.title}
                        value={insight.id}/>
                </label>
            )}
        </Form>
    )
}
