import PencilIcon from 'mdi-react/PencilIcon'
import React, { useCallback, useState } from 'react'
import { Link } from 'react-router-dom'
import { ExtensionsControllerNotificationProps } from '../../../../../shared/src/extensions/controller'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { CampaignDeleteButton } from '../common/CampaignDeleteButton'
import { EditCampaignForm } from '../common/EditCampaignForm'
import { CampaignsIcon } from '../icons'

interface Props extends ExtensionsControllerNotificationProps {
    campaign: GQL.ICampaign

    /** Called when the campaign is updated. */
    onCampaignUpdate: () => void
}

/**
 * A row in the list of campaigns.
 */
export const CampaignRow: React.FunctionComponent<Props> = ({ campaign, onCampaignUpdate, ...props }) => {
    const [isEditing, setIsEditing] = useState(false)
    const toggleIsEditing = useCallback(() => setIsEditing(!isEditing), [isEditing])

    return isEditing ? (
        <EditCampaignForm campaign={campaign} onCampaignUpdate={onCampaignUpdate} onDismiss={toggleIsEditing} />
    ) : (
        <div className="d-flex align-items-center justify-content-between">
            <h3 className="mb-0">
                <Link to={campaign.url} className="text-decoration-none">
                    <CampaignsIcon className="icon-inline" /> {campaign.name}
                </Link>
            </h3>
            <div className="text-right">
                <button type="button" className="btn btn-link text-decoration-none" onClick={toggleIsEditing}>
                    <PencilIcon className="icon-inline" /> Edit
                </button>
                <CampaignDeleteButton {...props} campaign={campaign} onDelete={onCampaignUpdate} />
            </div>
        </div>
    )
}
