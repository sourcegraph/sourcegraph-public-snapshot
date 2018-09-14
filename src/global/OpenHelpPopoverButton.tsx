import HelpCircleOutlineIcon from 'mdi-react/HelpCircleOutlineIcon'
import * as React from 'react'
import { eventLogger } from '../tracking/eventLogger'

function logEvent(): void {
    eventLogger.log('OpenHelpPopoverButtonClicked')
}

/** A button that, when clicked, toggles the visibility of the help popover. */
export const OpenHelpPopoverButton: React.SFC<{
    className?: string
    text?: string
    onHelpPopoverToggle: () => void
}> = ({ className = '', text, onHelpPopoverToggle }) => (
    <button
        className={`btn btn-link open-help-popover-button ${className}`}
        title={text ? undefined : 'Help and issues'}
        type="button"
        onClick={onHelpPopoverToggle}
        onMouseDown={logEvent}
    >
        <HelpCircleOutlineIcon className="icon-inline" />
        {text && <span className="open-help-popover-button__text ml-1">{text}</span>}
    </button>
)
