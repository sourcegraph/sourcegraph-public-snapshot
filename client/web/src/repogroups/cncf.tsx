import React from 'react'

import { RepogroupPage, RepogroupPageProps } from './RepogroupPage'
import { RepogroupMetadata } from './types'

export const cncf: RepogroupMetadata = {
    title: 'Cloud Native Computing Foundation (CNCF)',
    name: 'cncf',
    url: '/cncf',
    description: (
        <>
            Search the{' '}
            <a href="https://landscape.cncf.io/project=hosted" target="_blank" rel="noopener noreferrer">
                CNCF projects
            </a>
        </>
    ),
    examples: [],
    homepageDescription: 'Search CNCF projects.',
    homepageIcon: 'https://raw.githubusercontent.com/cncf/artwork/master/other/cncf/icon/color/cncf-icon-color.png',
    lowProfile: true,
}

export const CncfRepogroupPage: React.FunctionComponent<Omit<RepogroupPageProps, 'repogroupMetadata'>> = props => (
    <RepogroupPage {...props} repogroupMetadata={cncf} />
)
