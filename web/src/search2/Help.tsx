import HelpIcon from '@sourcegraph/icons/lib/Help'
import * as React from 'react'

export const Help: React.SFC = () => (
    <a className="search2-help" href="https://about.sourcegraph.com/docs/search" target="_blank">
        <HelpIcon className="icon-inline" />
        Help
    </a>
)
