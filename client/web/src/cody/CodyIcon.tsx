import React from 'react'

import { Icon } from '@sourcegraph/wildcard'

export const codyIconPath =
    'm9 15a1 1 0 01-1-1v-2a1 1 0 012 0v2a1 1 0 01-1 1zm6 0a1 1 0 01-1-1v-2a1 1 0 012 0v2a1 1 0 01-1 1zm-9-7a1 1 0 01-.71-.29l-3-3a1 1 0 011.42-1.42l3 3a1 1 0 010 1.42 1 1 0 01-.71.29zm12 0a1 1 0 01-.71-.29 1 1 0 010-1.42l3-3a1 1 0 111.42 1.42l-3 3a1 1 0 01-.71.29zm3 12h-18a1 1 0 01-1-1v-4.5a10 10 0 0120 0v4.5a1 1 0 01-1 1zm-17-2h16v-3.5a8 8 0 00-16 0z'

export const CodyIcon: React.FunctionComponent<{ className?: string }> = ({ className }) => (
    <Icon svgPath={codyIconPath} className={className} aria-hidden={true} />
)
