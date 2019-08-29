import React, { useState, useCallback } from 'react'
import { CampaignFormData } from '../../form/CampaignForm'
import { ButtonDropdown, DropdownToggle, DropdownMenu, DropdownItem } from 'reactstrap'
import { CheckableDropdownItem } from '../../../../components/CheckableDropdownItem'
import { Timestamp } from '../../../../components/time/Timestamp'
import { isAfter } from 'date-fns/esm'

interface Props {
    value: Pick<CampaignFormData, 'draft' | 'startDate'>
    onChange: (value: Pick<CampaignFormData, 'draft'>) => void

    icon?: React.ComponentType<{ className?: string }>

    disabled?: boolean
    className?: string
}

export const CreateCampaignButton: React.FunctionComponent<Props> = ({
    value,
    onChange,
    icon: Icon,
    disabled,
    className = '',
}) => {
    const [isOpen, setIsOpen] = useState(false)
    const toggleIsOpen = useCallback(() => setIsOpen(!isOpen), [isOpen])

    const onCreateClick = useCallback(() => onChange({ draft: false }), [onChange])
    const onDraftClick = useCallback(() => onChange({ draft: true }), [onChange])

    return (
        <ButtonDropdown isOpen={isOpen} toggle={toggleIsOpen} disabled={disabled}>
            <button type="submit" disabled={disabled} className={className}>
                {Icon && <Icon className="icon-inline" />}{' '}
                {value.draft ? 'Draft' : value.startDate ? 'Schedule' : 'Create'} campaign
            </button>
            <DropdownToggle
                disabled={disabled}
                color=""
                className={`create-campaign-button__dropdown-toggle pl-2 pr-3 ${className}`}
                caret={true}
            />
            <DropdownMenu className="create-campaign-button__dropdown">
                <CheckableDropdownItem
                    checked={!value.draft}
                    onClick={onCreateClick}
                    className="create-campaign-button__dropdown-item"
                >
                    <h3 className="font-weight-bold mb-0">Create {value.startDate && 'scheduled'} campaign</h3>
                    <p className="mb-0">
                        {value.startDate && isAfter(new Date(value.startDate), new Date()) && (
                            <>
                                Starting <Timestamp date={value.startDate} />:{' '}
                            </>
                        )}
                        Automatically create branches, issues, pull requests, and notifications{' '}
                    </p>
                </CheckableDropdownItem>
                <DropdownItem divider={true} />
                <CheckableDropdownItem
                    checked={value.draft}
                    onClick={onDraftClick}
                    className="create-campaign-button__dropdown-item"
                >
                    <h3 className="font-weight-bold mb-0">Create draft campaign</h3>
                    <p className="mb-0">Don't create branches, issues, pull requests, or notifications</p>
                </CheckableDropdownItem>
            </DropdownMenu>
        </ButtonDropdown>
    )
}
