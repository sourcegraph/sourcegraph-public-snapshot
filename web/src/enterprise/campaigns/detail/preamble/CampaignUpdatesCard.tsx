import React from 'react'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { CampaignChangesetsEditButton } from '../changesets/CampaignChangesetsEditButton'
import FileMultipleOutlineIcon from 'mdi-react/FileMultipleOutlineIcon'
import FileUploadOutlineIcon from 'mdi-react/FileUploadOutlineIcon'
import { Link } from 'react-router-dom'
import { Timestamp } from '../../../../components/time/Timestamp'
import H from 'history'
import { CampaignChangesetsAddExistingButton } from '../changesets/CampaignChangesetsAddExistingButton'
import { CampaignsIcon } from '../../icons'
import MenuDownIcon from 'mdi-react/MenuDownIcon'
import EyeIcon from 'mdi-react/EyeIcon'

interface Props {
    campaign: Pick<GQL.ICampaign, 'id' | 'url' | 'viewerCanAdminister' | 'closedAt'> & /* TODO(sqs) */ {
        patchesSetAt: string | null
        patchSetter: Pick<GQL.IUser, 'username' | 'url'> | null
        changesets: Pick<GQL.ICampaign['changesets'], 'totalCount'>
    }
    history: H.History
    className?: string
}

/**
 * A summary of the campaign's last patch update (if any) and contextual update actions that can be
 * performed, shown in the campaign preamble "timeline".
 */
export const CampaignUpdatesCard: React.FunctionComponent<Props> = ({ campaign, history, className = '' }) => {
    const isCampaignEmpty = campaign.patchesSetAt === null && campaign.changesets.totalCount === 0
    const Icon = isCampaignEmpty ? CampaignsIcon : campaign.patchesSetAt ? FileMultipleOutlineIcon : EyeIcon
    return (
        <div className={`card ${className} ${isCampaignEmpty ? 'border-primary' : ''}`}>
            <div className="card-body d-flex align-items-center">
                <Icon className={`h3 mb-0 mr-2 icon-inline ${isCampaignEmpty ? 'text-primary' : 'text-muted'}`} />
                <div className="d-flex align-items-center flex-1">
                    {campaign.patchesSetAt ? (
                        <span>
                            Campaign plan set{' '}
                            {campaign.patchSetter && (
                                <>
                                    by{' '}
                                    <Link to={campaign.patchSetter.url}>
                                        <strong>{campaign.patchSetter.username}</strong>
                                    </Link>
                                </>
                            )}{' '}
                            <span className="text-muted">
                                <Timestamp date={campaign.patchesSetAt} />
                            </span>
                        </span>
                    ) : isCampaignEmpty ? (
                        <strong>Start by adding changesets to this campaign.</strong>
                    ) : (
                        <span>This campaign is tracking existing changesets.</span>
                    )}
                    <div className="flex-1" />
                    {campaign.viewerCanAdminister && (
                        <>
                            {campaign.patchesSetAt === null && (
                                <CampaignChangesetsAddExistingButton
                                    campaign={campaign}
                                    history={history}
                                    buttonClassName="btn btn-secondary ml-3 pr-1"
                                />
                            )}
                            <CampaignChangesetsEditButton
                                campaign={campaign}
                                buttonClassName={`btn ${isCampaignEmpty ? 'btn-primary' : 'btn-secondary'} ml-3 pr-1`}
                            >
                                {campaign.patchesSetAt === null ? (
                                    <>
                                        Add plan <MenuDownIcon className="icon-inline" />
                                    </>
                                ) : undefined}
                            </CampaignChangesetsEditButton>
                        </>
                    )}
                </div>
            </div>{' '}
        </div>
    )
}
