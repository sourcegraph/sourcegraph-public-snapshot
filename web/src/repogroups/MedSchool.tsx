import { RepogroupMetadata } from './types'
import { SearchPatternType } from '../../../shared/src/graphql/schema'
import * as React from 'react'

export const stanfordmed: RepogroupMetadata = {
    title: 'Stanford Medicine',
    name: 'stanfordmed',
    url: '/stanford_medicine',
    description:
        'Use these search examples to explore Stanford Medicine repositories.',
    examples: [
        {
            title: 'Search all Stanford School of Medicine repositories.',
            exampleQuery: (
                <>
                    <span className="repogroup-page__keyword-text">repo:</span>/susom/{' '}
                </>
            ),
            patternType: SearchPatternType.literal,
            rawQuery: 'repo:/susom/',
        },
        {
            title: 'Find all School of Medicine repos written in a specific language.',
            exampleQuery: (
                <>repo:/susom/ lang:php</>
            ),
            patternType: SearchPatternType.literal,
            rawQuery: 'repo:/susom/ lang:php',
        },
        {
            title: 'Search a repository for all uses of a specific function. ',
            exampleQuery: (
                <>
                    repo:susom/database ConfigFromImpl()
                </>
            ),
            patternType: SearchPatternType.literal,
            rawQuery: 'repo:susom/database ConfigFromImpl()',
        },
    ],
    homepageDescription: 'Explore Stanford School of Medicine repositories.',
    homepageIcon: 'https://www.clipartmax.com/png/middle/170-1707817_stanford-medicine-logo-stanford-medicine-logo.png',
}
