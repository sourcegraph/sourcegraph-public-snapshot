import React from 'react'

import { SearchPatternType } from '@sourcegraph/shared/src/graphql-operations'

import { RepogroupPage, RepogroupPageProps } from './RepogroupPage'
import { RepogroupMetadata } from './types'

export const chakraui: RepogroupMetadata = {
    title: 'CHAKRA UI',
    name: 'chakraui',
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

export const ChakraUIRepogroupPage: React.FunctionComponent<Omit<RepogroupPageProps, 'repogroupMetadata'>> = props => (
    <RepogroupPage {...props} repogroupMetadata={chakraui} />
)
