import React from 'react'

import { SearchPatternType } from '../graphql-operations'

import { CommunitySearchContextPage, type CommunitySearchContextPageProps } from './CommunitySearchContextPage'
import type { CommunitySearchContextMetadata } from './types'

export const julia: CommunitySearchContextMetadata = {
    title: 'Julia',
    spec: 'julia',
    url: '/julia',
    description: 'Explore over 20 different Julia repositories. Search with examples below.',
    examples: [
        {
            title: "List all TODO's in Julia code",
            query: 'lang:Julia TODO case:yes',
            patternType: SearchPatternType.standard,
        },
        {
            title: 'Browse diffs for recent code changes',
            query: 'type:diff after:"1 week ago"',
            patternType: SearchPatternType.standard,
        },
    ],
    homepageDescription: 'Search within the Julia community.',
    homepageIcon: 'https://raw.githubusercontent.com/sourcegraph-community/julia-context/main/julia.svg',
}

export const JuliaCommunitySearchContextPage: React.FunctionComponent<
    React.PropsWithChildren<Omit<CommunitySearchContextPageProps, 'communitySearchContextMetadata'>>
> = props => <CommunitySearchContextPage {...props} communitySearchContextMetadata={julia} />
