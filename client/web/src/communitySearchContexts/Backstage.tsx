import React from 'react'

import { SearchPatternType } from '../graphql-operations'

import { CommunitySearchContextPage, type CommunitySearchContextPageProps } from './CommunitySearchContextPage'
import type { CommunitySearchContextMetadata } from './types'

export const backstage: CommunitySearchContextMetadata = {
    title: 'Backstage',
    spec: 'backstage',
    url: '/backstage',
    description: 'Explore over 25 different Backstage related repositories. Search with examples below.',
    examples: [
        {
            title: 'Browse diffs for recent code changes',
            query: 'type:diff after:"1 week ago"',
            patternType: SearchPatternType.standard,
        },
    ],
    homepageDescription: 'Search within the Backstage community.',
    homepageIcon: 'https://raw.githubusercontent.com/sourcegraph-community/backstage-context/main/backstage.svg',
}

export const BackstageCommunitySearchContextPage: React.FunctionComponent<
    React.PropsWithChildren<Omit<CommunitySearchContextPageProps, 'communitySearchContextMetadata'>>
> = props => <CommunitySearchContextPage {...props} communitySearchContextMetadata={backstage} />
