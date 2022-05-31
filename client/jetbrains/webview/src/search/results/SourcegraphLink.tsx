import React from 'react'

import { Link } from '@sourcegraph/wildcard'

import { openSourcegraphUrlInBrowser } from '../js-to-java-bridge'

interface Props {
    relativeUrl: string
    children: React.ReactNode
}

function onClick(event: React.MouseEvent<HTMLAnchorElement>, relativeUrl: string): void {
    event.preventDefault()
    event.stopPropagation()
    openSourcegraphUrlInBrowser(relativeUrl)
        .then(() => {})
        .catch(() => {})
}

export const SourcegraphLink: React.FunctionComponent<Props> = ({ relativeUrl, children }: Props) => (
    <Link to={relativeUrl} onClick={event => onClick(event, relativeUrl)}>
        {children}
    </Link>
)
