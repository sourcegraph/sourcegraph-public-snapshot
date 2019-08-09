import H from 'history'
import React from 'react'
import { Link } from 'react-router-dom'
import { CampaignFormControl } from './CampaignForm'
import { CAMPAIGN_TEMPLATES } from './templates'

interface Props extends CampaignFormControl {
    urlToFormWithTemplate: (templateID: string) => H.LocationDescriptor

    className?: string
}

export const CampaignTemplateChooser: React.FunctionComponent<Props> = ({ urlToFormWithTemplate, className = '' }) => {
    const a = 1
    console.log(a)
    return (
        <div className={className}>
            <div className="card">
                <ul className="list-group list-group-flush">
                    {CAMPAIGN_TEMPLATES.map((template, i) => (
                        <li
                            key={i}
                            className="list-group-item position-relative d-flex align-items-start justify-content-between"
                        >
                            <div>
                                <h5 className="mb-0">{template.title}</h5>
                                {template.detail && <p className="mb-0">{template.detail}</p>}
                            </div>
                            <Link to={urlToFormWithTemplate(template.id)} className="btn btn-primary">
                                Get started
                            </Link>
                        </li>
                    ))}
                </ul>
            </div>
        </div>
    )
}
