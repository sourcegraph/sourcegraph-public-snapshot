import { NotificationType } from '@sourcegraph/extension-api-classes'
import CloseIcon from 'mdi-react/CloseIcon'
import React, { useCallback } from 'react'
import { Link } from 'react-router-dom'
import { map, mapTo } from 'rxjs/operators'
import { displayRepoName } from '../../../../../../shared/src/components/RepoFileLink'
import { ExtensionsControllerNotificationProps } from '../../../../../../shared/src/extensions/controller'
import { dataOrThrowErrors, gql } from '../../../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { mutateGraphQL } from '../../../../backend/graphql'
import { ThreadStatusFields } from '../../../threadlike/threadStatus/threadStatus'
import { ThreadStatusIcon } from '../../../threadlike/threadStatus/ThreadStatusIcon'
import GithubCircleIcon from 'mdi-react/GithubCircleIcon'

const removeThreadsFromCampaign = (input: GQL.IRemoveThreadsFromCampaignOnMutationArguments): Promise<void> =>
    mutateGraphQL(
        gql`
            mutation RemoveThreadsFromCampaign($campaign: ID!, $threads: [ID!]!) {
                removeThreadsFromCampaign(campaign: $campaign, threads: $threads) {
                    alwaysNil
                }
            }
        `,
        input
    )
        .pipe(
            map(dataOrThrowErrors),
            mapTo(void 0)
        )
        .toPromise()

interface Props extends ExtensionsControllerNotificationProps {
    campaign: Pick<GQL.ICampaign, 'id'>
    thread: Pick<GQL.ThreadOrIssueOrChangeset, 'id' | 'repository' | 'title' | 'url' | 'externalURL'> &
        ThreadStatusFields
    onUpdate: () => void
}

/**
 * An item in the list of a campaign's threads.
 */
export const CampaignThreadListItem: React.FunctionComponent<Props> = ({
    campaign,
    thread,
    onUpdate,
    extensionsController,
}) => {
    const onRemoveClick = useCallback(async () => {
        try {
            await removeThreadsFromCampaign({ campaign: campaign.id, threads: [thread.id] })
            onUpdate()
        } catch (err) {
            extensionsController.services.notifications.showMessages.next({
                message: `Error removing thread from campaign: ${err.message}`,
                type: NotificationType.Error,
            })
        }
    }, [campaign.id, extensionsController.services.notifications.showMessages, onUpdate, thread.id])

    return (
        <div className="d-flex align-items-center">
            <Link to={thread.url} className="text-decoration-none flex-1">
                <ThreadStatusIcon thread={thread} className="mr-2" />
                <span className="text-muted mr-2">{displayRepoName(thread.repository.name)}:</span>
                {thread.title}
            </Link>
            {thread.externalURL && (
                <a href={thread.externalURL}>
                    <GithubCircleIcon className="icon-inline mr-1" />
                </a>
            )}
            <button
                className="btn btn-link btn-sm p-1"
                aria-label="Remove thread from campaign"
                onClick={onRemoveClick}
            >
                <CloseIcon className="icon-inline" />
            </button>
        </div>
    )
}
