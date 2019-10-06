import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import SyncIcon from 'mdi-react/SyncIcon'
import React, { useCallback, useState } from 'react'
import { map, mapTo } from 'rxjs/operators'
import { NotificationType } from '../../../../../shared/src/api/client/services/notifications'
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'
import { dataOrThrowErrors, gql } from '../../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { mutateGraphQL, queryGraphQL } from '../../../backend/graphql'
import { getCompleteCampaignExtensionData } from '../extensionData'
import { Workflow } from '../../../schema/workflow.schema'

const queryCampaignWorkflowInfo = (
    campaign: Pick<GQL.IExpCampaign, 'id'>
): Promise<{ workflow: Workflow; name: string; body?: string }> =>
    queryGraphQL(
        gql`
            query CampaignWorkflowInfo($campaign: ID!) {
                node(id: $campaign) {
                    __typename
                    ... on ExpCampaign {
                        name
                        body
                        workflow {
                            parsed
                        }
                    }
                }
            }
        `,
        { campaign: campaign.id }
    )
        .pipe(
            map(dataOrThrowErrors),
            map(data => {
                if (!data.node || data.node.__typename !== 'ExpCampaign') {
                    throw new Error('invalid campaign')
                }
                return {
                    name: data.node.name,
                    body: data.node.body,
                    workflow: data.node.workflow ? data.node.workflow.parsed : {},
                }
            })
        )
        .toPromise()

const forceRefreshCampaign = (args: GQL.IForceRefreshCampaignOnMutationArguments): Promise<void> =>
    mutateGraphQL(
        gql`
            mutation ForceRefreshCampaign($campaign: ID!, $extensionData: ExpCampaignExtensionData!) {
                forceRefreshCampaign(campaign: $campaign, extensionData: $extensionData) {
                    id
                }
            }
        `,
        args
    )
        .pipe(
            map(dataOrThrowErrors),
            mapTo(undefined)
        )
        .toPromise()

interface Props extends ExtensionsControllerProps {
    campaign: Pick<GQL.IExpCampaign, 'id'>
    onComplete?: () => void
    className?: string
    buttonClassName?: string
}

/**
 * A button that force-refreshes a campaign.
 */
export const CampaignForceRefreshButton: React.FunctionComponent<Props> = ({
    campaign,
    onComplete,
    className = '',
    buttonClassName = 'btn-link text-decoration-none',
    extensionsController,
}) => {
    const [isLoading, setIsLoading] = useState(false)
    const onClick = useCallback<React.FormEventHandler>(
        async e => {
            e.preventDefault()
            setIsLoading(true)
            try {
                const { workflow, name, body } = await queryCampaignWorkflowInfo(campaign)
                const extensionData = await getCompleteCampaignExtensionData(extensionsController, workflow, {
                    name,
                    body,
                })
                await forceRefreshCampaign({
                    campaign: campaign.id,
                    extensionData,
                })
                setIsLoading(false)
                if (onComplete) {
                    onComplete()
                }
            } catch (err) {
                setIsLoading(false)
                extensionsController.services.notifications.showMessages.next({
                    message: `Error force-refreshing campaign: ${err.message}`,
                    type: NotificationType.Error,
                })
            }
        },
        [campaign, extensionsController, onComplete]
    )
    return (
        <button type="button" disabled={isLoading} className={`btn ${buttonClassName} ${className}`} onClick={onClick}>
            {isLoading ? <LoadingSpinner className="icon-inline" /> : <SyncIcon className="icon-inline" />} Refresh
        </button>
    )
}
