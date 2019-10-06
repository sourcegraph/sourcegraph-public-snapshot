import { NotificationType } from '@sourcegraph/extension-api-classes'
import H from 'history'
import React, { useCallback, useState } from 'react'
import { map } from 'rxjs/operators'
import { ExtensionsControllerProps } from '../../../../../../shared/src/extensions/controller'
import { dataOrThrowErrors, gql } from '../../../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { PlatformContextProps } from '../../../../../../shared/src/platform/context'
import { mutateGraphQL } from '../../../../backend/graphql'
import { ThemeProps } from '../../../../theme'
import { getCompleteCampaignExtensionData } from '../../extensionData'
import { CampaignFormData } from '../../form/CampaignForm'
import { CampaignUpdatePreview } from '../../updatePreview/CampaignUpdatePreview'
import { CampaignAreaContext } from '../CampaignArea'
import { EditCampaignForm } from './EditCampaignForm'
import { Workflow } from '../../../../schema/workflow.schema'
import { parseJSON } from '../../../../settings/configuration'

export const updateCampaign = (input: GQL.IExpUpdateCampaignInput): Promise<GQL.IExpCampaign> =>
    mutateGraphQL(
        gql`
            mutation UpdateCampaign($input: ExpUpdateCampaignInput!) {
                expUpdateCampaign(input: $input) {
                    id
                    url
                }
            }
        `,
        { input }
    )
        .pipe(
            map(dataOrThrowErrors),
            map(data => data.expUpdateCampaign)
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
        startDate: campaign.startDate,
        dueDate: campaign.dueDate,
        draft: campaign.isDraft,
        workflowAsJSONCString: campaign.workflow ? campaign.workflow.raw : '',
    })
    const onChange = useCallback((newValue: Partial<CampaignFormData>) => {
        setValue((prevValue: CampaignFormData) => ({ ...prevValue, ...newValue }))
    }, [])

    const [isLoading, setIsLoading] = useState(false)
    const onSubmit = useCallback(async () => {
        setIsLoading(true)
        try {
            const extensionData = await getCompleteCampaignExtensionData(
                props.extensionsController,
                parseJSON(value.workflowAsJSONCString) as Workflow,
                value
            )
            await updateCampaign({
                id: campaign.id,
                ...value,
                clearStartDate: value.startDate === undefined,
                clearDueDate: value.dueDate === undefined,
                extensionData,
            })
            setIsLoading(false)
        } catch (err) {
            setIsLoading(false)
            props.extensionsController.services.notifications.showMessages.next({
                message: `Error updating campaign: ${err.message}`,
                type: NotificationType.Error,
            })
            throw err
        }
    }, [campaign.id, props.extensionsController, value])

    return (
        <div className={`campaign-manage-page ${className}`}>
            <EditCampaignForm
                {...props}
                value={value}
                isValid={true /* TODO!(sqs) */}
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
