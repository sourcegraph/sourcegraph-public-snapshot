import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import H from 'history'
import React from 'react'
import { CreateCampaignButton } from './CreateCampaignButton'
import { CampaignFormProps, CampaignForm } from '../../form/CampaignForm'
import { WorkflowEditor } from '../../form/WorkflowEditor'
import { ThemeProps } from '../../../../theme'

interface Props extends CampaignFormProps, ThemeProps {
    location: H.Location
    history: H.History
}

/**
 * A form to create a new campaign.
 */
export const NewCampaignForm: React.FunctionComponent<Props> = props => (
    <CampaignForm {...props}>
        {({ form }) => (
            <>
                {form}
                <WorkflowEditor {...props} />
                <div className="form-group mt-4">
                    <CreateCampaignButton
                        {...props}
                        icon={props.isLoading ? LoadingSpinner : undefined}
                        className="btn btn-success"
                        disabled={!props.isValid || props.disabled || props.isLoading}
                    />
                </div>
            </>
        )}
    </CampaignForm>
)
