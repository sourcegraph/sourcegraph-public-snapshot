import H from 'history'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import React, { useState, useCallback } from 'react'
import { ExtensionsControllerProps } from '../../../../../../shared/src/extensions/controller'
import { PlatformContextProps } from '../../../../../../shared/src/platform/context'
import { ThemeProps } from '../../../../theme'
import { CampaignAreaContext } from '../CampaignArea'
import { PageTitle } from '../../../../components/PageTitle'
import { CampaignFormData } from '../../form/CampaignForm'
import { EditCampaignForm } from './EditCampaignForm'
import { dataOrThrowErrors, gql } from '../../../../../../shared/src/graphql/graphql'
import { map, first } from 'rxjs/operators'
import { mutateGraphQL } from '../../../../backend/graphql'
import { getCampaignExtensionData } from '../../extensionData'
import { RuleDefinition } from '../../../rules/types'
import { NotificationType } from '@sourcegraph/extension-api-classes'
import { CampaignUpdatePreview } from '../../updatePreview/CampaignUpdatePreview'

export const updateCampaign = (input: GQL.IUpdateCampaignInput): Promise<GQL.ICampaign> =>
    mutateGraphQL(
        gql`
            mutation UpdateCampaign($input: UpdateCampaignInput!) {
                updateCampaign(input: $input) {
                    id
                    url
                }
            }
        `,
        { input }
    )
        .pipe(
            map(dataOrThrowErrors),
            map(data => data.updateCampaign)
        )
        .toPromise()

interface Props
    extends Pick<CampaignAreaContext, 'campaign'>,
        ExtensionsControllerProps,
        PlatformContextProps,
        ThemeProps {
    className?: string

    /**
     * The base URL of the area.
     */
    match: { url: string }

    location: H.Location
    history: H.History
}

export const CampaignManagePage: React.FunctionComponent<Props> = ({ campaign, className = '', ...props }) => {
    const [value, setValue] = useState<CampaignFormData>({
        namespace: campaign.namespace.id,
        name: campaign.name,
        body: campaign.body,
        template: campaign.template
            ? {
                  template: campaign.template.template,
                  context: campaign.template.context.parsed /* TODO!(sqs): preserve jsonc */,
              }
            : undefined,
        startDate: campaign.startDate,
        dueDate: campaign.dueDate,
        draft: campaign.isDraft,
        isValid: true,
        rules: campaign.rules.nodes.map(
            rule =>
                ({
                    name: rule.name,
                    description: rule.description,
                    definition: rule.definition.raw,
                } as GQL.INewRuleInput)
        ),
    })
    const onChange = useCallback((newValue: Partial<CampaignFormData>) => {
        setValue((prevValue: CampaignFormData) => ({ ...prevValue, ...newValue }))
    }, [])

    const [isLoading, setIsLoading] = useState(false)
    const onSubmit = useCallback(async () => {
        setIsLoading(true)
        try {
            const extensionData = await getCampaignExtensionData(
                props.extensionsController,
                value.rules ? value.rules.map(rule => JSON.parse(rule.definition) as RuleDefinition) : []
            )
                .pipe(first())
                .toPromise()
            await updateCampaign({
                id: campaign.id,
                ...value,
                clearTemplate: value.template === undefined,
                clearStartDate: value.startDate === undefined,
                clearDueDate: value.dueDate === undefined,
                extensionData,
            })
        } catch (err) {
            setIsLoading(false)
            props.extensionsController.services.notifications.showMessages.next({
                message: `Error updating campaign: ${err.message}`,
                type: NotificationType.Error,
            })
        }
    }, [campaign.id, props.extensionsController, value])

    return (
        <div className={`campaign-manage-page ${className}`}>
            <EditCampaignForm
                {...props}
                value={value}
                onChange={onChange}
                onSubmit={onSubmit}
                isLoading={isLoading}
                className="flex-1"
            />
            <hr className="my-5" />
            <CampaignUpdatePreview {...props} campaign={campaign} data={value} className="mb-5" />
        </div>
    )
}
