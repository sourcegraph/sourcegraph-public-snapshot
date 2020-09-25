import React from 'react'

export interface CampaignCloseHeaderProps {
    // Nothing.
}

const CampaignCloseHeader: React.FunctionComponent<CampaignCloseHeaderProps> = () => (
    <>
        <span />
        <h5 className="text-uppercase text-center text-nowrap">Action</h5>
        <h5 className="text-uppercase text-nowrap">Changeset information</h5>
        <h5 className="text-uppercase text-center text-nowrap">Check state</h5>
        <h5 className="text-uppercase text-center text-nowrap">Review state</h5>
        <h5 className="text-uppercase text-center text-nowrap">Changes</h5>
    </>
)

export const CampaignCloseHeaderWillCloseChangesets: React.FunctionComponent<CampaignCloseHeaderProps> = () => (
    <>
        <h2 className="campaign-close-header__row test-campaigns-close-willclose-header">
            Closing the campaign will close the following changesets:
        </h2>
        <CampaignCloseHeader />
    </>
)

export const CampaignCloseHeaderWillKeepChangesets: React.FunctionComponent<CampaignCloseHeaderProps> = () => (
    <>
        <h2 className="campaign-close-header__row">The following changesets will remain open:</h2>
        <CampaignCloseHeader />
    </>
)
