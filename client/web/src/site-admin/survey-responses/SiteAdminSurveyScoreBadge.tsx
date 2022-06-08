import React from 'react'

import { Badge } from '@sourcegraph/wildcard'

import { scoreToClassSuffix } from './utils'

export const ScoreBadge: React.FunctionComponent<React.PropsWithChildren<{ score: number }>> = props => (
    <Badge className="ml-4" pill={true} variant={scoreToClassSuffix(props.score)} tooltip={`${props.score} out of 10`}>
        Score: {props.score}
    </Badge>
)
