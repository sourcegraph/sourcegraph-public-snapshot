import HelpIcon from '@sourcegraph/icons/lib/Help'
import * as React from 'react'
import { eventLogger } from '../tracking/eventLogger'

const onDidClickSearchHelp = (): void => {
    eventLogger.log('SearchHelpButtonClicked')
}

export const SearchHelp: React.SFC = () => (
    <a
        onClick={onDidClickSearchHelp}
        className="search-help"
        href="https://about.sourcegraph.com/docs/search"
        target="_blank"
        data-tooltip="View search documentation"
    >
        <small>
            <HelpIcon className="icon-inline" />
            <span className="search-help__text">Help</span>
        </small>
    </a>
)
