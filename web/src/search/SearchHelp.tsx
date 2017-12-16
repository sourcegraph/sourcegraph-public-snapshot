import HelpIcon from '@sourcegraph/icons/lib/Help'
import * as React from 'react'

export const SearchHelp: React.SFC = () => (
    <a className="search-help" href="https://about.sourcegraph.com/docs/search" target="_blank">
        <small>
            <HelpIcon className="icon-inline" />
            <span className="search-help__text">Help</span>
        </small>
    </a>
)
