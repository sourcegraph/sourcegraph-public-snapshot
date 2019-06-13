import React from 'react'
import { CampaignTemplate } from './templates'

interface Props {
    template: CampaignTemplate

    tag?: 'li'
    after?: React.ReactFragment
    className?: string
}

export const CampaignTemplateRow: React.FunctionComponent<Props> = ({
    template: { icon: Icon, ...template },
    tag: Tag = 'li',
    after,
    className = '',
}) => (
    <Tag className={`d-flex align-items-start justify-content-between ${className}`}>
        <Icon className="icon-inline flex-0 h2 mb-0 mr-3" />
        <div className="flex-1 mr-3">
            <h4 className="mb-0">{template.title}</h4>
            {template.detail && <p className="mb-0 text-muted">{template.detail}</p>}
        </div>
        {after}
    </Tag>
)
