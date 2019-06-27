import H from 'history'
import React from 'react'
import { ExtensionsControllerProps } from '../../../../../../shared/src/extensions/controller'
import { Checklist } from '../../checklist'
import { ChecklistDescription } from './ChecklistDescription'

interface Props extends ExtensionsControllerProps {
    checklist: Checklist

    areaURL: string

    className?: string
    history: H.History
    location: H.Location
}

/**
 * The overview for a single checklist.
 */
export const ChecklistOverview: React.FunctionComponent<Props> = ({ checklist, areaURL, className = '', ...props }) => (
    <div className={`checklist-overview ${className || ''}`}>
        <ChecklistDescription {...props} checklist={checklist} />
    </div>
)
