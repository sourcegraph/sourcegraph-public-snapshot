import React from 'react'
import { Link } from 'react-router-dom'
import { ErrorLike, isErrorLike } from '../../../../../../shared/src/util/errors'
import { RuleTemplate } from '../../form/templates'
import { Markdown } from '../../../../../../shared/src/components/Markdown'
import { renderMarkdown } from '../../../../../../shared/src/util/markdown'

interface Props {
    template: null | RuleTemplate | ErrorLike
}

export const NewCampaignRuleTemplateFormGroupHeader: React.FunctionComponent<Props> = ({ template }) => {
    const TemplateIcon = template !== null && !isErrorLike(template) ? template.icon : undefined

    return template === null || isErrorLike(template) ? (
        <>{isErrorLike(template) && <div className="alert alert-danger">{template.message}</div>}</>
    ) : (
        <>
            {template.isEmpty ? (
                <h2>New campaign</h2>
            ) : (
                <>
                    <h2 className="d-flex align-items-start">
                        {TemplateIcon && <TemplateIcon className="icon-inline mr-2 flex-0" />} New campaign:{' '}
                        {template.title}
                    </h2>
                    <p>
                        {template.detail && (
                            <Markdown dangerousInnerHTML={renderMarkdown(template.detail)} inline={true} />
                        )}{' '}
                        <Link to="?" className="text-muted mb-2">
                            Choose a different template.
                        </Link>
                    </p>
                </>
            )}
        </>
    )
}
