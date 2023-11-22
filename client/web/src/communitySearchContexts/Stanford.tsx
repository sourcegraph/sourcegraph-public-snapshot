import React from 'react'

import { SearchPatternType } from '../graphql-operations'

import { CommunitySearchContextPage, type CommunitySearchContextPageProps } from './CommunitySearchContextPage'
import type { CommunitySearchContextMetadata } from './types'

export const stanford: CommunitySearchContextMetadata = {
    title: 'Stanford University',
    spec: 'stanford',
    url: '/stanford',
    description: 'Explore open-source code from Stanford students, faculty, research groups, and clubs.',
    examples: [
        {
            title: 'Find all mentions of "machine learning" in Stanford projects.',
            patternType: SearchPatternType.standard,
            query: 'machine learning',
        },
        {
            title: 'Explore the code of specific research groups like Hazy Research, a group that investigates machine learning models and automated training set creation.',
            patternType: SearchPatternType.standard,
            query: 'repo:/HazyResearch/',
        },
        {
            title: 'Explore the code of a specific user or organization such as Stanford University School of Medicine.',
            patternType: SearchPatternType.standard,
            query: 'repo:/susom/',
        },
        {
            title: 'Search for repositories related to introductory programming concepts.',
            patternType: SearchPatternType.standard,
            query: 'repo:recursion',
        },
        {
            title: 'Explore the README files of thousands of projects.',
            patternType: SearchPatternType.standard,
            query: 'file:README.txt',
        },
        {
            title: 'Find old-style string formatted print statements.',
            patternType: SearchPatternType.structural,
            query: 'lang:python print(:[args] % :[v])',
        },
    ],
    homepageDescription: 'Explore Stanford open-source code.',
    homepageIcon:
        'https://upload.wikimedia.org/wikipedia/commons/thumb/a/aa/Icons8_flat_graduation_cap.svg/120px-Icons8_flat_graduation_cap.svg.png',
}

export const StanfordCommunitySearchContextPage: React.FunctionComponent<
    React.PropsWithChildren<Omit<CommunitySearchContextPageProps, 'communitySearchContextMetadata'>>
> = props => <CommunitySearchContextPage {...props} communitySearchContextMetadata={stanford} />
