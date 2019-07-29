import SlackIcon from 'mdi-react/SlackIcon'
import React, { useCallback, useState } from 'react'
import { ButtonDropdown, DropdownToggle } from 'reactstrap'
import { SlackMessageDropdownMenu } from './SlackMessageDropdownMenu'

interface Props {
    buttonClassName?: string
}

/**
 * An action on a thread inbox item to open a Slack message with somebody involved with the code.
 */
export const ThreadInboxItemSlackMessage: React.FunctionComponent<Props> = ({ buttonClassName = 'btn-secondary' }) => {
    const [isOpen, setIsOpen] = useState(false)
    const toggleIsOpen = useCallback(() => setIsOpen(!isOpen), [isOpen])

    return (
        <ButtonDropdown isOpen={isOpen} toggle={toggleIsOpen}>
            <DropdownToggle className={`btn ${buttonClassName}`} color="none">
                <SlackIcon className="icon-inline" /> Slack message...
            </DropdownToggle>
            <SlackMessageDropdownMenu right={true} />
        </ButtonDropdown>
    )
}
