import React from 'react'

import { SearchPatternType } from '@sourcegraph/shared/src/graphql-operations'

import { CommunitySearchContextPage, type CommunitySearchContextPageProps } from './CommunitySearchContextPage'
import type { CommunitySearchContextMetadata } from './types'

export const o3de: CommunitySearchContextMetadata = {
    title: 'O3DE',
    spec: 'o3de',
    url: '/o3de',
    description: '',
    examples: [
        {
            title: 'Search for O3DE gems',
            patternType: SearchPatternType.standard,
            query: 'file:gem.json',
        },
        {
            title: 'Browse diffs for recent code changes',
            patternType: SearchPatternType.standard,
            query: 'type:diff after:"1 week ago"',
        },
    ],
    homepageDescription: 'Search within the O3DE organization.',
    homepageIcon:
        'https://raw.githubusercontent.com/o3de/artwork/19b89e72e15824f20204a8977a007f53d5fcd5b8/o3de/03_O3DE%20Application%20Icon/SVG/O3DE-Circle-Icon.svg',
}

export const O3deCommunitySearchContextPage: React.FunctionComponent<
    React.PropsWithChildren<Omit<CommunitySearchContextPageProps, 'communitySearchContextMetadata'>>
> = props => <CommunitySearchContextPage {...props} communitySearchContextMetadata={o3de} />
