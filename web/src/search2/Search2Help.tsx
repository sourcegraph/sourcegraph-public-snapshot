import HelpIcon from '@sourcegraph/icons/lib/Help'
import * as React from 'react'

export const Search2Help: React.SFC = () => (
    <a className="search2-help" href="https://about.sourcegraph.com/docs/search" target="_blank">
        <small>
            <HelpIcon className="icon-inline" />
            <span>Help</span>
        </small>
    </a>
)
