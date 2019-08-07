import { NotificationType } from '@sourcegraph/extension-api-classes'
import CheckIcon from 'mdi-react/CheckIcon'
import CloseIcon from 'mdi-react/CloseIcon'
import GithubCircleIcon from 'mdi-react/GithubCircleIcon'
import React, { useCallback } from 'react'
import { Link } from 'react-router-dom'
import { map, mapTo } from 'rxjs/operators'
import { displayRepoName } from '../../../../../../shared/src/components/RepoFileLink'
import { ExtensionsControllerNotificationProps } from '../../../../../../shared/src/extensions/controller'
import { dataOrThrowErrors, gql } from '../../../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { mutateGraphQL } from '../../../../backend/graphql'
import { ThreadStateFields } from '../../../threads/common/threadState/threadState'
import { ThreadStateIcon } from '../../../threads/common/threadState/ThreadStateIcon'

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
    thread: Pick<GQL.IThread, 'id' | 'repository' | 'title' | 'url'> & ThreadStateFields
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
            <Link to={thread.url} className="text-decoration-none">
                <ThreadStateIcon thread={thread} className="mr-2" />
                <span className="text-muted mr-2">{displayRepoName(thread.repository.name)}:</span>
                {thread.title}
            </Link>
            {thread.repository.name.length % 5 === 0 /* TODO!(sqs) */ ? (
                <>
                    <CheckIcon className="ml-2 mr-1 icon-inline text-success" />
                    All checks pass
                </>
            ) : (
                <>
                    <CloseIcon className="ml-2 mr-1 icon-inline text-danger" />1 failing check
                </>
            )}

            <div className="flex-1" />
            {thread.repository.name.length % 3 === 0 /* TODO!(sqs) */ ? (
                <span className="badge badge-primary mr-2">Changes requested</span>
            ) : (
                <span className="badge badge-success mr-2">Approved</span>
            )}
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
