import React from 'react'

import { SearchPatternType } from '@sourcegraph/shared/src/graphql-operations'

import { CommunitySearchContextPage, CommunitySearchContextPageProps } from './CommunitySearchContextPage'
import { CommunitySearchContextMetadata } from './types'

export const chakraui: CommunitySearchContextMetadata = {
    title: 'CHAKRA UI',
    spec: 'chakraui',
    url: '/chakraui',
    description: '',
    examples: [
        {
            title: 'Search for Chakra UI packages',
            patternType: SearchPatternType.literal,
            query: 'file:package.json',
        },
        {
            title: 'Browse diffs for recent code changes',
            patternType: SearchPatternType.literal,
            query: 'type:diff after:"1 week ago"',
        },
    ],
    homepageDescription: 'Search within the Chakra UI organization.',
    homepageIcon: 'https://raw.githubusercontent.com/chakra-ui/chakra-ui/main/logo/logomark-colored.svg',
}

export const ChakraUICommunitySearchContextPage: React.FunctionComponent<
    React.PropsWithChildren<Omit<CommunitySearchContextPageProps, 'communitySearchContextMetadata'>>
> = props => <CommunitySearchContextPage {...props} communitySearchContextMetadata={chakraui} />
