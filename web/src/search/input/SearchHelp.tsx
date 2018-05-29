import HelpIcon from '@sourcegraph/icons/lib/Help'
import * as React from 'react'
import { eventLogger } from '../../tracking/eventLogger'

const onDidClickSearchHelp = (): void => {
    eventLogger.log('SearchHelpButtonClicked')
}

export const SearchHelp: React.SFC<{ className?: string; compact?: boolean }> = ({
    className = '',
    compact = false,
}) => (
    <a
        onClick={onDidClickSearchHelp}
        className={`search-help ${className}`}
        href="https://about.sourcegraph.com/docs/search/query-syntax"
        target="_blank"
        title="View search documentation"
    >
        <HelpIcon className="icon-inline" /> {!compact && <span className="search-help__text">Help</span>}
    </a>
)
