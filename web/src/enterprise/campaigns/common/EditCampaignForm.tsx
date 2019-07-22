import React, { useCallback, useState } from 'react'
import { map } from 'rxjs/operators'
import { gql } from '../../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { createAggregateError } from '../../../../../shared/src/util/errors'
import { mutateGraphQL } from '../../../backend/graphql'
import { CampaignForm, CampaignFormData } from '../form/CampaignForm'

const updateCampaign = (input: GQL.IUpdateCampaignInput): Promise<void> =>
    mutateGraphQL(
        gql`
            mutation UpdateCampaign($input: UpdateCampaignInput!) {
                campaigns {
                    updateCampaign(input: $input) {
                        id
                    }
                }
            }
        `,
        { input }
    )
        .pipe(
            map(({ data, errors }) => {
                if (!data || !data.campaigns || !data.campaigns.updateCampaign || (errors && errors.length > 0)) {
                    throw createAggregateError(errors)
                }
            })
        )
        .toPromise()

interface Props {
    campaign: Pick<GQL.ICampaign, 'id'> & CampaignFormData

    /** Called when the form is dismissed. */
    onDismiss: () => void

    /** Called after the campaign is updated successfully. */
    onCampaignUpdate: () => void

    className?: string
}

/**
 * A form to edit a campaign.
 */
export const EditCampaignForm: React.FunctionComponent<Props> = ({
    campaign,
    onDismiss,
    onCampaignUpdate,
    className = '',
}) => {
    const [isLoading, setIsLoading] = useState(false)
    const onSubmit = useCallback(
        async ({ name }: CampaignFormData) => {
            setIsLoading(true)
            try {
                await updateCampaign({ id: campaign.id, name })
                setIsLoading(false)
                onDismiss()
                onCampaignUpdate()
            } catch (err) {
                setIsLoading(false)
                alert(err.message) // TODO!(sqs)
            }
        },
        [onDismiss, onCampaignUpdate, campaign.id]
    )

    return (
        <CampaignForm
            initialValue={campaign}
            onDismiss={onDismiss}
            onSubmit={onSubmit}
            buttonText="Save changes"
            isLoading={isLoading}
            className={className}
        />
    )
}
