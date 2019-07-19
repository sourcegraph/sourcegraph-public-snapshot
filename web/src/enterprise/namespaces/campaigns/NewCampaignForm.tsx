import React, { useCallback, useState } from 'react'
import { map } from 'rxjs/operators'
import { gql } from '../../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { createAggregateError } from '../../../../../shared/src/util/errors'
import { mutateGraphQL } from '../../../backend/graphql'
import { NamespaceAreaContext } from '../../../namespaces/NamespaceArea'
import { CampaignForm, CampaignFormData } from './CampaignForm'

const createCampaign = (input: GQL.ICreateCampaignInput): Promise<void> =>
    mutateGraphQL(
        gql`
            mutation CreateCampaign($input: CreateCampaignInput!) {
                campaigns {
                    createCampaign(input: $input) {
                        id
                    }
                }
            }
        `,
        { input }
    )
        .pipe(
            map(({ data, errors }) => {
                if (!data || !data.campaigns || !data.campaigns.createCampaign || (errors && errors.length > 0)) {
                    throw createAggregateError(errors)
                }
            })
        )
        .toPromise()

interface Props extends Pick<NamespaceAreaContext, 'namespace'> {
    /** Called when the form is dismissed. */
    onDismiss: () => void

    /** Called after the campaign is created successfully. */
    onCampaignCreate: () => void

    className?: string
}

/**
 * A form to create a new campaign.
 */
export const NewCampaignForm: React.FunctionComponent<Props> = ({
    namespace,
    onDismiss,
    onCampaignCreate,
    className = '',
}) => {
    const [isLoading, setIsLoading] = useState(false)
    const onSubmit = useCallback(
        async ({ name }: CampaignFormData) => {
            setIsLoading(true)
            try {
                await createCampaign({ name, namespace: namespace.id })
                setIsLoading(false)
                onDismiss()
                onCampaignCreate()
            } catch (err) {
                setIsLoading(false)
                alert(err.message) // TODO!(sqs)
            }
        },
        [namespace.id, onDismiss, onCampaignCreate]
    )

    return (
        <CampaignForm
            onDismiss={onDismiss}
            onSubmit={onSubmit}
            buttonText="Create campaign"
            isLoading={isLoading}
            className={className}
        />
    )
}
