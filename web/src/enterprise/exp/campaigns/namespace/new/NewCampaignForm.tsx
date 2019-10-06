import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import H from 'history'
import React, { useCallback } from 'react'
import { CreateCampaignButton } from './CreateCampaignButton'
import { CampaignFormProps, CampaignForm } from '../../form/CampaignForm'
import { RuleTemplateFormGroup } from '../../form/templates/RuleTemplateFormGroup'
import { NewCampaignRuleTemplateFormGroupHeader } from './NewCampaignRuleTemplateFormGroupHeader'
import { Link } from 'react-router-dom'
import { RuleTemplateChooser } from '../../form/RuleTemplateChooser'
import { EMPTY_RULE_TEMPLATE_ID } from '../../form/templates'

interface Props extends CampaignFormProps {
    /** The URL of the form. */
    match: {
        url: string
    }
    location: H.Location
}

/**
 * A form to create a new campaign.
 */
export const NewCampaignForm: React.FunctionComponent<Props> = ({ match, location, ...props }) => {
    const urlToFormWithTemplate = useCallback(
        (templateID: string) => `${match.url}?${new URLSearchParams({ template: templateID }).toString()}`,
        [match.url]
    )

    return (
        <CampaignForm {...props}>
            {({ form }) =>
                !props.value.rules ? (
                    <>
                        <h2 className="mb-3">New campaign</h2>
                        <RuleTemplateChooser urlToFormWithTemplate={urlToFormWithTemplate} className="mb-4" />
                        <p>
                            Don't see what you're looking for?{' '}
                            <Link to={urlToFormWithTemplate(EMPTY_RULE_TEMPLATE_ID)}>Create a new empty campaign</Link>{' '}
                            and manually add changesets, issues, and rules.
                        </p>
                    </>
                ) : (
                    <>
                        {props.value.rules.map((_rule, i) => (
                            <RuleTemplateFormGroup
                                // eslint-disable-next-line react/no-array-index-key
                                key={i}
                                {...props}
                                ruleIndex={i}
                                url={match.url}
                                header={NewCampaignRuleTemplateFormGroupHeader}
                            />
                        ))}
                        {form}
                        <div className="form-group mt-4">
                            <CreateCampaignButton
                                {...props}
                                icon={props.isLoading ? LoadingSpinner : undefined}
                                className="btn btn-success"
                                disabled={!props.value.isValid || props.disabled || props.isLoading}
                            />
                        </div>
                    </>
                )
            }
        </CampaignForm>
    )
}
