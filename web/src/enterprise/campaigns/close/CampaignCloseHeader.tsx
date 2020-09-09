import React from 'react'

export interface CampaignCloseHeaderProps {
    // Nothing.
}

export const CampaignCloseHeader: React.FunctionComponent<CampaignCloseHeaderProps> = () => (
    <>
        <span />
        <h5 className="text-uppercase text-center text-nowrap">Action</h5>
        <h5 className="text-uppercase text-nowrap">Changeset information</h5>
        <h5 className="text-uppercase text-center text-nowrap">Check state</h5>
        <h5 className="text-uppercase text-center text-nowrap">Review state</h5>
        <h5 className="text-uppercase text-center text-nowrap">Changes</h5>
    </>
)
