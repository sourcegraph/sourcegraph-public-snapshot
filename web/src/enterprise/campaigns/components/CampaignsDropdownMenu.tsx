import React from 'react'
import { DropdownItem, DropdownMenu } from 'reactstrap'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { isErrorLike } from '../../../../../shared/src/util/errors'
import { useCampaigns } from '../list/useCampaigns'

interface Props {
    /** Called when the user selects a campaign in the menu. */
    onSelect: (campaign: Pick<GQL.ICampaign, 'id'>) => void
}

const LOADING = 'loading' as const

/**
 * A dropdown menu with a list of campaigns.
 */
export const CampaignsDropdownMenu: React.FunctionComponent<Props> = ({ onSelect, ...props }) => {
    const campaigns = useCampaigns()
    return (
        <DropdownMenu {...props}>
            {campaigns === LOADING ? (
                <DropdownItem header={true} className="py-1">
                    Loading campaigns...
                </DropdownItem>
            ) : isErrorLike(campaigns) ? (
                <DropdownItem header={true} className="py-1">
                    Error loading campaigns
                </DropdownItem>
            ) : campaigns.nodes.length === 0 ? (
                <DropdownItem header={true}>No campaigns exist</DropdownItem>
            ) : (
                campaigns.nodes.map(campaign => (
                    // tslint:disable-next-line: jsx-no-lambda
                    <DropdownItem key={campaign.id} onClick={() => onSelect(campaign)}>
                        <small className="text-muted">#{campaign.namespace.namespaceName}</small> {campaign.name}
                    </DropdownItem>
                ))
            )}
        </DropdownMenu>
    )
}
