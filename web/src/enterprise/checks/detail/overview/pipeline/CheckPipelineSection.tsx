import BellIcon from 'mdi-react/BellIcon'
import SettingsIcon from 'mdi-react/SettingsIcon'
import React from 'react'
import * as sourcegraph from 'sourcegraph'
import { Markdown } from '../../../../../../../shared/src/components/Markdown'
import { renderMarkdown } from '../../../../../../../shared/src/util/markdown'

type CheckSection = keyof Required<sourcegraph.CheckInformation>['sections']

interface Props {
    section: CheckSection
    content: sourcegraph.MarkupContent
    action?: (className: string) => JSX.Element | null
}

const checkSectionIcon = (section: CheckSection): React.ComponentType<{ className?: string }> | null => {
    switch (section) {
        case 'settings':
            return SettingsIcon
        case 'notifications':
            return BellIcon
        default:
            return null
    }
}

/**
 * A section (step) in the pipeline for a check.
 */
export const CheckPipelineSection: React.FunctionComponent<Props> = ({ section, content, action }) => {
    const Icon = checkSectionIcon(section)
    return (
        <div className="check-pipeline-section d-flex align-items-start bg-body border p-3 mb-5">
            {Icon && <Icon className="icon-inline mr-3 mt-2 flex-0" />}
            <Markdown className="mt-2 flex-1" dangerousInnerHTML={renderMarkdown(content.value)}></Markdown>
            {action && action('btn-sm btn-secondary flex-0')}
        </div>
    )
}
