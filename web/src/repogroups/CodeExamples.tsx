import { RepogroupMetadata } from './types'
import { SearchPatternType } from '../../../shared/src/graphql/schema'
import * as React from 'react'

export const codeexamples: RepogroupMetadata = {
    title: 'Code Examples',
    name: 'codeexamples',
    url: '/code_examples',
    description:
        'The following examples showcase ways in which the Stanford Sourcegraph instance can help students learn to code.',
    examples: [
        {
            title: 'Search for repositories related to challenging concepts.',
            exampleQuery: (
                <>
                    repo:recursion
                </>
            ),
            patternType: SearchPatternType.literal,
            rawQuery: 'repo:recursion',
        },
        {
            title: 'Explore <i>Cracking the Coding Interview</i> solutions.',
            exampleQuery: (
                <>repo:alexhagiopol/cracking-the-coding-interview</>
            ),
            patternType: SearchPatternType.literal,
            rawQuery: 'repo:alexhagiopol/cracking-the-coding-interview',
        },
        {
            title: 'Explore the Stanford C++ Libraries.',
            exampleQuery: (
                <>repo:zelenski/stanford-cpp-library</>
            ),
            patternType: SearchPatternType.literal,
            rawQuery: 'repo:zelenski/stanford-cpp-library',
        },
    ],
    homepageDescription: 'Find helpful coding examples when trying to understand difficult concepts or practicing for an interview.',
    homepageIcon: 'https://web.stanford.edu/class/archive/cs/cs106b/cs106b.1164/img/overview/overview.png',
}
