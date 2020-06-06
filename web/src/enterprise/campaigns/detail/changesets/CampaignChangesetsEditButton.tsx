import React, { useState, useCallback } from 'react'
import MenuDownIcon from 'mdi-react/MenuDownIcon'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { Popover } from 'reactstrap'
import HelpCircleOutlineIcon from 'mdi-react/HelpCircleOutlineIcon'

interface Props {
    campaign: Pick<GQL.ICampaign, 'id' | 'url'>
    buttonClassName?: string
}

export const CampaignChangesetsEditButton: React.FunctionComponent<Props> = ({
    campaign,
    children = (
        <>
            Update plan <MenuDownIcon className="icon-inline" />
        </>
    ),
    buttonClassName = '',
}) => {
    const [isOpen, setIsOpen] = useState(false)
    const toggleIsOpen = useCallback<React.MouseEventHandler>(
        event => {
            event.preventDefault()
            setIsOpen(!isOpen)
        },
        [isOpen]
    )

    const [popoverTarget, setPopoverTarget] = useState<HTMLElement | null>(null)
    const popoverTargetReference = useCallback((element: HTMLElement | null) => setPopoverTarget(element), [])

    return (
        <>
            <button
                type="button"
                onClick={toggleIsOpen}
                className={`d-inline-flex align-items-center ${buttonClassName}`}
                ref={popoverTargetReference}
            >
                {children}
            </button>
            {popoverTarget && (
                <Popover
                    placement="bottom-end"
                    isOpen={isOpen}
                    target={popoverTarget}
                    toggle={toggleIsOpen}
                    innerClassName="p-3"
                    style={{ maxWidth: '54rem' }}
                >
                    <p>Generate and preview a new campaign plan:</p>
                    <pre style={{ backgroundColor: '#f3f3fa', overflow: 'auto' }} className="p-3">
                        <code>
                            src campaign update -preview -action action.json -id={campaign.id.replace(/=*$/, '')}
                        </code>
                    </pre>
                    <ul className="list-unstyled mb-0 mt-3">
                        <li>
                            <a href="TODO(sqs)" className="d-flex align-items-center">
                                <HelpCircleOutlineIcon className="icon-inline mr-1" /> How to install Sourcegraph CLI
                            </a>
                        </li>
                        <li>
                            <a href="TODO(sqs)" className="d-flex align-items-center">
                                <HelpCircleOutlineIcon className="icon-inline mr-1" /> How to define campaign actions
                            </a>
                        </li>
                    </ul>
                </Popover>
            )}
        </>
    )
}
