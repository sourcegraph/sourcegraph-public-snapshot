import H from 'history'
import React, { useCallback, useState } from 'react'
import { map, mapTo } from 'rxjs/operators'
import { NotificationType } from '../../../../../shared/src/api/client/services/notifications'
import { ExtensionsControllerNotificationProps } from '../../../../../shared/src/extensions/controller'
import { dataOrThrowErrors, gql } from '../../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { mutateGraphQL } from '../../../backend/graphql'
import { Form } from '../../../components/Form'

const updateCampaign = (args: GQL.IUpdateCampaignOnMutationArguments): Promise<void> =>
    mutateGraphQL(
        gql`
            mutation UpdateCampaign($input: UpdateCampaignInput!) {
                updateCampaign(input: $input) {
                    id
                }
            }
        `,
        args
    )
        .pipe(
            map(dataOrThrowErrors),
            mapTo(void 0)
        )
        .toPromise()

const publishPreviewCampaign = (args: GQL.IPublishPreviewCampaignOnMutationArguments): Promise<void> =>
    mutateGraphQL(
        gql`
            mutation PublishPreviewCampaign($campaign: ID!) {
                publishPreviewCampaign(campaign: $campaign) {
                    id
                }
            }
        `,
        args
    )
        .pipe(
            map(dataOrThrowErrors),
            mapTo(void 0)
        )
        .toPromise()

interface Props extends ExtensionsControllerNotificationProps {
    campaign: Pick<GQL.ICampaign, 'id' | 'name' | 'url'>
    onCampaignUpdate: () => void

    className?: string
    history: H.History
}

/**
 * A form to publish a preview changeset (to make it non-preview).
 */
export const CreateCampaignFromPreviewForm: React.FunctionComponent<Props> = ({
    campaign,
    onCampaignUpdate,
    className = '',
    history,
    extensionsController,
}) => {
    const [isLoading, setIsLoading] = useState(false)

    const [uncommittedName, setUncommittedTitle] = useState(campaign.name)
    const onChangeTitle = useCallback<React.ChangeEventHandler<HTMLInputElement>>(e => {
        setUncommittedTitle(e.currentTarget.value)
    }, [])

    const onSubmit: React.FormEventHandler = useCallback(
        async e => {
            e.preventDefault()
            setIsLoading(true)
            try {
                if (uncommittedName !== campaign.name) {
                    await updateCampaign({
                        input: {
                            id: campaign.id,
                            name: uncommittedName,
                        },
                    })
                }
                await publishPreviewCampaign({ campaign: campaign.id })
                setIsLoading(false)
                onCampaignUpdate()
                history.push(campaign.url) // TODO!(sqs): use Redirect component
            } catch (err) {
                setIsLoading(false)
                extensionsController.services.notifications.showMessages.next({
                    message: `Error publishing preview campaign: ${err.message}`,
                    type: NotificationType.Error,
                })
            }
        },
        [
            uncommittedName,
            campaign.name,
            campaign.id,
            campaign.url,
            onCampaignUpdate,
            history,
            extensionsController.services.notifications.showMessages,
        ]
    )

    return (
        <Form className={className} onSubmit={onSubmit}>
            <div className="form-group">
                <input
                    type="text"
                    className="form-control"
                    value={uncommittedName}
                    onChange={onChangeTitle}
                    placeholder="Title"
                    autoComplete="off"
                    autoFocus={true}
                    disabled={isLoading}
                />
            </div>
            <div className="form-group d-flex">
                <textarea
                    className="form-control"
                    onChange={e => {
                        // TODO!(sqs)
                        alert('not implemented')
                    }}
                    placeholder="Description"
                    style={{ resize: 'vertical', minHeight: '150px' }}
                    disabled={isLoading}
                />
            </div>
            <div className="d-flex justify-content-end">
                <button type="submit" className="btn btn-lg btn-success" disabled={isLoading}>
                    Create changeset
                </button>
            </div>
        </Form>
    )
}
