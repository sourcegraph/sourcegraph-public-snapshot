import React from 'react'

export const CampaignStateBadge: React.FunctionComponent<{ isClosed: boolean }> = ({ isClosed }) => {
    if (isClosed) {
        return <span className="badge badge-danger text-uppercase">Closed</span>
    }
    return <span className="badge badge-success text-uppercase">Open</span>
}
