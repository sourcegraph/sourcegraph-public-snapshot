import React from 'react'
import { ChangesetAreaContext } from './ChangesetArea'

interface Props extends Pick<ChangesetAreaContext, 'changeset'> {
    className?: string
}

/**
 * The overview for a single changeset.
 */
export const ChangesetOverview: React.FunctionComponent<Props> = ({ changeset, className = '' }) => (
    <div className={`changeset-overview ${className || ''}`}>
        <h2>{changeset.title}</h2>
    </div>
)
