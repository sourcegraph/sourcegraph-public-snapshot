import React from 'react'

import { SearchPatternType } from '@sourcegraph/shared/src/graphql-operations'

import { CommunitySearchContextPage, type CommunitySearchContextPageProps } from './CommunitySearchContextPage'
import type { CommunitySearchContextMetadata } from './types'

export const stackStorm: CommunitySearchContextMetadata = {
    title: 'StackStorm',
    spec: 'stackstorm',
    url: '/stackstorm',
    description: '',
    examples: [
        {
            title: 'Passive sensor examples',
            patternType: SearchPatternType.standard,
            query: 'from st2reactor.sensor.base import Sensor',
        },
        {
            title: 'Polling sensor examples',
            patternType: SearchPatternType.standard,
            query: 'from st2reactor.sensor.base import PollingSensor',
        },
        {
            title: 'Trigger examples in rules',
            patternType: SearchPatternType.standard,
            query: 'repo:Exchange trigger: file:.yaml$',
        },
        {
            title: 'Actions that use the Orquesta runner',
            patternType: SearchPatternType.regexp,
            query: 'repo:Exchange runner_type:\\s*"orquesta"',
        },
        {
            title: 'All instances where a trigger is injected with an explicit payload',
            patternType: SearchPatternType.structural,
            query: 'repo:Exchange sensor_service.dispatch(:[1], payload=:[2])',
        },
    ],
    homepageDescription: 'Search within the StackStorm and StackStorm Exchange community.',
    homepageIcon: 'https://avatars.githubusercontent.com/u/4969009?s=200&v=4',
}

export const StackStormCommunitySearchContextPage: React.FunctionComponent<
    React.PropsWithChildren<Omit<CommunitySearchContextPageProps, 'communitySearchContextMetadata'>>
> = props => <CommunitySearchContextPage {...props} communitySearchContextMetadata={stackStorm} />
