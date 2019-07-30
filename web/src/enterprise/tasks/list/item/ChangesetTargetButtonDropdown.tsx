import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import React, { useCallback, useState } from 'react'
import { ButtonDropdown, DropdownItem, DropdownMenu, DropdownToggle } from 'reactstrap'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { isErrorLike } from '../../../../../../shared/src/util/errors'
import { CheckableDropdownItem } from '../../../../components/CheckableDropdownItem'
import { CampaignsIcon } from '../../../campaigns/icons'
import { useCampaigns } from '../../../campaigns/list/useCampaigns'

export interface CreateOrPreviewChangesetButtonProps {
    onClick: () => void

    disabled?: boolean
    loading?: boolean
    className?: string
    buttonClassName?: string
}

const LOADING: 'loading' = 'loading'

/**
 * A button to preview a changeset or append to an existing changeset.
 */
export const ChangesetTargetButtonDropdown: React.FunctionComponent<CreateOrPreviewChangesetButtonProps> = ({
    onClick,
    disabled,
    loading,
    className = '',
    buttonClassName = '',
}) => {
    const [isOpen, setIsOpen] = useState(false)
    const toggleIsOpen = useCallback(() => setIsOpen(!isOpen), [isOpen])

    const campaigns = useCampaigns() // TODO!(sqs): filter to only relevant (open, changeset) campaigns

    const [appendToExistingCampaign, setAppendToExistingCampaign] = useState<Pick<GQL.ICampaign, 'id' | 'name'>>()
    const clearAppendToExistingCampaign = useCallback(() => setAppendToExistingCampaign(undefined), [])

    const Icon = loading ? LoadingSpinner : CampaignsIcon

    return (
        <ButtonDropdown
            isOpen={isOpen}
            toggle={toggleIsOpen}
            className={`changeset-target-button-dropdown ${className}`}
        >
            <button className={`btn ${buttonClassName}`} onClick={onClick} disabled={disabled}>
                <div
                    style={{
                        width: '22px' /* TODO!(sqs): avoid jitter bc loading spinner is not as wide as other icon */,
                    }}
                >
                    <Icon className="icon-inline mr-1" />
                </div>
                {appendToExistingCampaign === undefined ? 'New campaign' : 'Add to existing changeset'}
            </button>
            <DropdownToggle
                color="success"
                className="changeset-target-button-dropdown__dropdown-toggle pl-1 pr-2"
                caret={true}
                disabled={disabled}
            />
            <DropdownMenu>
                <CheckableDropdownItem
                    onClick={clearAppendToExistingCampaign}
                    checked={appendToExistingCampaign === undefined}
                >
                    <h5 className="mb-1">New changeset</h5>
                    <span className="text-muted">You can preview the changes before submitting</span>
                </CheckableDropdownItem>
                <DropdownItem divider={true} />
                {campaigns === LOADING ? (
                    <DropdownItem header={true} className="py-1">
                        Loading campaigns...
                    </DropdownItem>
                ) : isErrorLike(campaigns) ? (
                    <DropdownItem header={true} className="py-1">
                        Error loading campaigns
                    </DropdownItem>
                ) : (
                    <>
                        <DropdownItem header={true} className="py-1">
                            Add to existing campaign...
                        </DropdownItem>
                        {campaigns.nodes
                            .filter(c => !!c.name)
                            .slice(/*TODO!(sqs)*/ 0, 7)
                            .map(campaign => (
                                <CheckableDropdownItem
                                    key={campaign.id}
                                    // tslint:disable-next-line: jsx-no-lambda
                                    onClick={() => setAppendToExistingCampaign(campaign)}
                                    checked={Boolean(
                                        appendToExistingCampaign && appendToExistingCampaign.id === campaign.id
                                    )}
                                >
                                    {campaign.name}
                                </CheckableDropdownItem>
                            ))}
                    </>
                )}
            </DropdownMenu>
        </ButtonDropdown>
    )
}
