import React from 'react'
import { ChangesetPlanProps } from '../useChangesetPlan'

interface Props extends ChangesetPlanProps {
    className?: string
}

/**
 * A bar displaying the changesets related to a set of diagnostics, plus a preview of and statistics
 * about a new changeset that is being created.
 */
export const DiagnosticsChangesetsBar: React.FunctionComponent<Props> = ({ changesetPlan, className = '' }) => {
    const a = 123
    return (
        <div className={`diagnostics-changesets-bar d-flex align-items-center py-2 px-3 w-100 border ${className}`}>
            {changesetPlan.operations.length > 0 ? (
                <pre style={{ maxHeight: '150px', overflow: 'auto', fontSize: '9px' }}>
                    <code>{JSON.stringify(changesetPlan, null, 2)}</code>
                </pre>
            ) : (
                <span className="text-muted">Select actions to start creating a new changeset...</span>
            )}
        </div>
    )
}
