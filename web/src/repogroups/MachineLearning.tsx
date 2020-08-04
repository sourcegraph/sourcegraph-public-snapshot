import { RepogroupMetadata } from './types'
import { SearchPatternType } from '../../../shared/src/graphql/schema'
import * as React from 'react'

export const machinelearning: RepogroupMetadata = {
    title: 'Machine Learning',
    name: 'machinelearning',
    url: '/machine_learning',
    description:
        'Use these search examples to explore Stanford machine learning projects.',
    examples: [
        {
            title: 'Search general machine learning projects.',
            exampleQuery: (
                <>
                    <span className="repogroup-page__keyword-text">machine learning</span>
                </>
            ),
            patternType: SearchPatternType.literal,
            rawQuery: 'machine learning',
        },
        {
            title: 'Search a specfic research group.',
            exampleQuery: (
                <>
                    <span className="repogroup-page__keyword-text">repo:</span>/HazyResearch/{' '}
                </>
            ),
            patternType: SearchPatternType.literal,
            rawQuery: 'repo:/HazyResearch/',
        },
        {
            title: 'Search a specfic machine learning project.',
            exampleQuery: (
                <>
                    <span className="repogroup-page__keyword-text">repo:</span>emmjaykay/stanford_self_driving_car_code{' '}
                </>
            ),
            patternType: SearchPatternType.literal,
            rawQuery: 'repo:emmjaykay/stanford_self_driving_car_code',
        },
    ],
    homepageDescription: 'Find Stanford machine learning projects.',
    homepageIcon: 'https://aicenter.stanford.edu/sites/g/files/sbiybj12571/f/styles/4-col-header/public/ai-horiz_rgb.jpg?itok=MSq5IQ9P',
}
