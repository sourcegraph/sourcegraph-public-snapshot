import React from 'react'

export interface CampaignChangesetsHeaderProps {
    // Nothing.
}

export const CampaignChangesetsHeader: React.FunctionComponent<CampaignChangesetsHeaderProps> = () => (
    <>
        <span />
        <h5 className="text-uppercase text-center text-nowrap">Status</h5>
        <h5 className="text-uppercase text-nowrap">Changeset information</h5>
        <h5 className="text-uppercase text-center text-nowrap">Check state</h5>
        <h5 className="text-uppercase text-center text-nowrap">Review state</h5>
        <h5 className="text-uppercase text-center text-nowrap">Changes</h5>
    </>
)
