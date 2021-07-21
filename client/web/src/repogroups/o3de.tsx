import React from 'react'

import { SearchPatternType } from '@sourcegraph/shared/src/graphql-operations'

import { RepogroupPage, RepogroupPageProps } from './RepogroupPage'
import { RepogroupMetadata } from './types'

export const o3de: RepogroupMetadata = {
    title: 'O3DE',
    name: 'o3de',
    url: '/o3de',
    description: '',
    examples: [
        {
            title: 'Search for O3DE gems',
            patternType: SearchPatternType.literal,
            query: 'file:gem.json',
        },
        {
            title: 'Browse diffs for recent code changes',
            patternType: SearchPatternType.literal,
            query: 'type:diff after:"1 week ago"',
        },
    ],
    homepageDescription: 'Search within the O3DE organization.',
    homepageIcon:
        'https://raw.githubusercontent.com/o3de/artwork/19b89e72e15824f20204a8977a007f53d5fcd5b8/o3de/03_O3DE%20Application%20Icon/SVG/O3DE-Circle-Icon.svg',
}

export const O3deRepogroupPage: React.FunctionComponent<Omit<RepogroupPageProps, 'repogroupMetadata'>> = props => (
    <RepogroupPage {...props} repogroupMetadata={o3de} />
)
