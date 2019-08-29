import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import React from 'react'
import { CreateCampaignButton } from './CreateCampaignButton'
import { CampaignFormProps, CampaignForm } from '../../form/CampaignForm'

interface Props extends CampaignFormProps {}

/**
 * A form to create a new campaign.
 */
export const NewCampaignForm: React.FunctionComponent<Props> = props => (
    <CampaignForm {...props}>
        {({ form }) => (
            <>
                <h2>New campaign</h2>
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
        )}
    </CampaignForm>
)
