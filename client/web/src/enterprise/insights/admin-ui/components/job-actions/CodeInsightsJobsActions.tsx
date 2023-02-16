import { ReactElement } from 'react'

import classNames from 'classnames'

import { Button, ButtonProps } from '@sourcegraph/wildcard'

import styles from './CodeInsightsJobsActions.module.scss'

interface CodeInsightsJobActionsProps {
    selectedJobIds: string[]
    className?: string
    onSelectionClear: () => void
}

export function CodeInsightsJobsActions(props: CodeInsightsJobActionsProps): ReactElement {
    const { selectedJobIds, className, onSelectionClear } = props

    return (
        <div className={classNames(className, styles.actions)}>
            <JobActionButton actionCount={selectedJobIds.length}>Retry</JobActionButton>
            <JobActionButton actionCount={selectedJobIds.length}>Front of queue</JobActionButton>
            <JobActionButton actionCount={selectedJobIds.length}>Back of queue</JobActionButton>
            {selectedJobIds.length > 0 && (
                <Button variant="secondary" outline={true} onClick={onSelectionClear}>
                    Clear selection
                </Button>
            )}
        </div>
    )
}

interface JobActionButton extends ButtonProps {
    actionCount: number
}

function JobActionButton(props: JobActionButton): ReactElement {
    const { actionCount, children } = props

    return (
        <Button {...props} variant="primary" disabled={actionCount === 0} className={styles.action}>
            {actionCount > 0 && <span className={styles.count}>{actionCount}</span>}
            {children}
        </Button>
    )
}
