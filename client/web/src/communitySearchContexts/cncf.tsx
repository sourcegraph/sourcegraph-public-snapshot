import React from 'react'

import { CommunitySearchContextPage, CommunitySearchContextPageProps } from './CommunitySearchContextPage'
import { CommunitySearchContextMetadata } from './types'

export const cncf: CommunitySearchContextMetadata = {
    title: 'Cloud Native Computing Foundation (CNCF)',
    spec: 'cncf',
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

export const CncfCommunitySearchContextPage: React.FunctionComponent<
    Omit<CommunitySearchContextPageProps, 'communitySearchContextMetadata'>
> = props => <CommunitySearchContextPage {...props} communitySearchContextMetadata={cncf} />
