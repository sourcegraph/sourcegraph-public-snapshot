import React from 'react'

interface Props {
    icon?: React.ComponentType<{ className?: string }>

    disabled?: boolean
    className?: string
}

export const CreateCampaignButton: React.FunctionComponent<Props> = ({ icon: Icon, disabled, className = '' }) => (
    <button type="submit" disabled={disabled} className={className}>
        {Icon && <Icon className="icon-inline" />} Create campaign
    </button>
)
