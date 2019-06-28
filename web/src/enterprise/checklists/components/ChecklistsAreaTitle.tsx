import React from 'react'
import { ChecklistIcon } from '../icons'

interface Props {
    className?: string
    primaryActions?: JSX.Element | null
}

/**
 * The checklist area title.
 *
 * // TODO!(sqs): dedupe with ChecksAreaTitle?
 */
export const ChecklistAreaTitle: React.FunctionComponent<Props> = ({ className = '', primaryActions, children }) => (
    <div className="d-flex align-items-center mb-3">
        <h1 className={`h3 mb-0 d-flex align-items-center ${className}`}>
            <ChecklistIcon className="icon-inline mr-1" /> Checklist
        </h1>
        {children}
        {primaryActions && (
            <>
                <div className="flex-1" />
                {primaryActions}
            </>
        )}
    </div>
)
