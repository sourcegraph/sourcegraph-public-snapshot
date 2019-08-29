import React, { useCallback, useState } from 'react'
import { map, mapTo } from 'rxjs/operators'
import { NotificationType } from '../../../../../../shared/src/api/client/services/notifications'
import { ExtensionsControllerNotificationProps } from '../../../../../../shared/src/extensions/controller'
import { dataOrThrowErrors, gql } from '../../../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { mutateGraphQL } from '../../../../backend/graphql'
import { Form } from '../../../../components/Form'

export const updateCampaign = (args: GQL.IUpdateCampaignOnMutationArguments): Promise<void> =>
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
            mapTo(undefined)
        )
        .toPromise()

interface Props extends ExtensionsControllerNotificationProps {
    campaign: Pick<GQL.ICampaign, 'id' | 'name'>
    onCampaignUpdate: () => void
    className?: string

    /** Provided by tests only. */
    _updateCampaign?: typeof updateCampaign
}

/**
 * The name in the campaign header, which has an edit mode.
 */
// tslint:disable: jsx-no-lambda
export const CampaignHeaderEditableName: React.FunctionComponent<Props> = ({
    campaign,
    onCampaignUpdate,
    className = '',
    extensionsController,
    _updateCampaign = updateCampaign,
}) => {
    const [state, setState] = useState<'viewing' | 'editing' | 'loading'>('viewing')
    const [uncommittedName, setUncommittedName] = useState(campaign.name)

    const onSubmit: React.FormEventHandler = async e => {
        e.preventDefault()
        setState('loading')
        try {
            await _updateCampaign({ input: { id: campaign.id, name: uncommittedName } })
            onCampaignUpdate()
        } catch (err) {
            extensionsController.services.notifications.showMessages.next({
                message: `Error editing campaign name: ${err.message}`,
                type: NotificationType.Error,
            })
        } finally {
            setState('viewing')
        }
    }

    const onEditClick = useCallback(() => setState('editing'), [])
    const onCancelClick = useCallback<React.MouseEventHandler>(
        e => {
            e.preventDefault()
            setState('viewing')
            setUncommittedName(campaign.name)
        },
        [campaign.name]
    )

    return state === 'viewing' ? (
        <div className={`d-flex align-items-start justify-content-between ${className}`}>
            <h1 className="font-weight-normal mb-0 h2 mr-2">{campaign.name}</h1>
            <button type="button" className="btn btn-secondary btn-sm mt-1" onClick={onEditClick}>
                Edit
            </button>
        </div>
    ) : (
        <Form className={`form d-flex ${className}`} onSubmit={onSubmit}>
            <input
                type="text"
                className="form-control flex-1 mr-2"
                value={uncommittedName}
                onChange={e => setUncommittedName(e.currentTarget.value)}
                placeholder="Title"
                autoComplete="off"
                autoFocus={true}
                disabled={state === 'loading'}
            />
            <div className="text-nowrap flex-0 d-flex align-items-center">
                <button type="submit" className="btn btn-success" disabled={state === 'loading'}>
                    Save
                </button>
                <button type="reset" className="btn btn-link" onClick={onCancelClick} disabled={state === 'loading'}>
                    Cancel
                </button>
            </div>
        </Form>
    )
}
