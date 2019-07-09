import BellIcon from 'mdi-react/BellIcon'
import SettingsIcon from 'mdi-react/SettingsIcon'
import React from 'react'
import * as sourcegraph from 'sourcegraph'
import { Markdown } from '../../../../../../../shared/src/components/Markdown'
import { renderMarkdown } from '../../../../../../../shared/src/util/markdown'

type StatusSection = keyof Required<sourcegraph.Status>['sections']

interface Props {
    section: StatusSection
    content: sourcegraph.MarkupContent
    action?: (className: string) => JSX.Element | null
}

const statusSectionIcon = (section: StatusSection): React.ComponentType<{ className?: string }> | null => {
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
 * A section (step) in the pipeline for a status.
 */
export const StatusPipelineSection: React.FunctionComponent<Props> = ({ section, content, action }) => {
    const Icon = statusSectionIcon(section)
    return (
        <div className="status-pipeline-section d-flex align-items-start bg-body border p-3 mb-5">
            {Icon && <Icon className="icon-inline mr-3 mt-2 flex-0" />}
            <Markdown className="mt-2 flex-1" dangerousInnerHTML={renderMarkdown(content.value)}></Markdown>
            {action && action('btn-sm btn-secondary flex-0')}
        </div>
    )
}
