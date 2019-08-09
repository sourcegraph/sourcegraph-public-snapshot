import H from 'history'
import React from 'react'
import { Link } from 'react-router-dom'
import { CampaignFormControl } from './CampaignForm'
import { CampaignTemplateRow } from './CampaignTemplateRow'
import { CAMPAIGN_TEMPLATES } from './templates'

interface Props extends CampaignFormControl {
    urlToFormWithTemplate: (templateID: string) => H.LocationDescriptor

    className?: string
}

export const CampaignTemplateChooser: React.FunctionComponent<Props> = ({ urlToFormWithTemplate, className = '' }) => (
    <div className={className}>
        <div className="card">
            <ul className="list-group list-group-flush">
                {CAMPAIGN_TEMPLATES.filter(template => !template.isEmpty).map((template, i) => (
                    <CampaignTemplateRow
                        key={i}
                        template={template}
                        after={
                            <Link
                                to={urlToFormWithTemplate(template.id)}
                                className="btn btn-primary flex-0 stretched-link mt-2 mr-1"
                            >
                                Get started
                            </Link>
                        }
                        className="list-group-item position-relative"
                    />
                ))}
            </ul>
        </div>
    </div>
)
