import React from 'react'
import { ChangesetIcon } from '../icons'

interface Props {
    className?: string
    primaryActions?: JSX.Element | null
}

/**
 * The changesets area title.
 *
 * // TODO!(sqs): dedupe with ChecksAreaTitle?
 */
export const ChangesetsAreaTitle: React.FunctionComponent<Props> = ({ className = '', primaryActions, children }) => (
    <div className="d-flex align-items-center mb-3">
        <h1 className={`h3 mb-0 d-flex align-items-center ${className}`}>
            <ChangesetIcon className="icon-inline mr-1" /> Changesets
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
