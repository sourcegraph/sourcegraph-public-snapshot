import { RepogroupMetadata } from './types'
import { SearchPatternType } from '../../../shared/src/graphql/schema'
import * as React from 'react'

export const clubsprojects: RepogroupMetadata = {
    title: 'Clubs and Projects',
    name: 'clubsprojects',
    url: '/clubs_projects',
    description:
        'Use these search examples to explore the code of Stanford clubs and open source projects.',
    examples: [
        {
            title: 'Explore the README files of thousands of projects.',
            exampleQuery: (
                <>
                    <span className="repogroup-page__keyword-text">file:</span>README.txt{' '}
                </>
            ),
            patternType: SearchPatternType.literal,
            rawQuery: 'file:README.txt',
        },
        {
            title: 'Perform a general search if you\'re unsure of the project name.',
            exampleQuery: (
                <>
                    space initiative
                </>
            ),
            patternType: SearchPatternType.literal,
            rawQuery: 'space initiative',
        },
        {
            title: 'Find Stanford projects in python.',
            exampleQuery: (
                <>
                    <span className="repogroup-page__keyword-text">lang:</span>python{' '}
                </>
            ),
            patternType: SearchPatternType.literal,
            rawQuery: 'lang:python',
        },
    ],
    homepageDescription: 'Find Stanford open source projects and club code.',
    homepageIcon: 'https://upload.wikimedia.org/wikipedia/en/thumb/b/b7/Stanford_University_seal_2003.svg/1200px-Stanford_University_seal_2003.svg.png',
}
