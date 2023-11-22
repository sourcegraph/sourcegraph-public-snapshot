import React from 'react'

import { Link } from '@sourcegraph/wildcard'

import { CommunitySearchContextPage, type CommunitySearchContextPageProps } from './CommunitySearchContextPage'
import type { CommunitySearchContextMetadata } from './types'

export const cncf: CommunitySearchContextMetadata = {
    title: 'Cloud Native Computing Foundation (CNCF)',
    spec: 'cncf',
    url: '/cncf',
    description: (
        <>
            Search the{' '}
            <Link to="https://landscape.cncf.io/project=hosted" target="_blank" rel="noopener noreferrer">
                CNCF projects
            </Link>
        </>
    ),
    examples: [],
    homepageDescription: 'Search CNCF projects.',
    homepageIcon: 'https://raw.githubusercontent.com/cncf/artwork/master/other/cncf/icon/color/cncf-icon-color.png',
    lowProfile: true,
}

export const CncfCommunitySearchContextPage: React.FunctionComponent<
    React.PropsWithChildren<Omit<CommunitySearchContextPageProps, 'communitySearchContextMetadata'>>
> = props => <CommunitySearchContextPage {...props} communitySearchContextMetadata={cncf} />
