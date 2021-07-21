import React from 'react'

import { SearchPatternType } from '@sourcegraph/shared/src/graphql-operations'

import { RepogroupPage, RepogroupPageProps } from './RepogroupPage'
import { RepogroupMetadata } from './types'

export const temporal: RepogroupMetadata = {
    title: 'Temporal',
    name: 'temporalio',
    url: '/temporal',
    description: '',
    examples: [
        {
            title: 'All test functions',
            patternType: SearchPatternType.literal,
            query: 'type:symbol Test',
        },
        {
            title: 'Search for a specifc function or class',
            patternType: SearchPatternType.literal,
            query: 'type:symbol SimpleSslContextBuilder',
        },
    ],
    homepageDescription: 'Search within the Temporal organization.',
    homepageIcon: 'https://avatars.githubusercontent.com/u/56493103?s=200&v=4',
}

export const TemporalRepogroupPage: React.FunctionComponent<Omit<RepogroupPageProps, 'repogroupMetadata'>> = props => (
    <RepogroupPage {...props} repogroupMetadata={temporal} />
)
