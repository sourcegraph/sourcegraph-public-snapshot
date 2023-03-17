import React from 'react'

import { H1 } from '@sourcegraph/wildcard'

import { CodyIcon } from '../../cody/CodyIcon'

import { CodyChat } from './CodyChat'

export const CodyPage: React.FunctionComponent<{}> = () => (
    <div>
        <H1>
            <CodyIcon /> Cody
        </H1>
        <CodyChat />
    </div>
)
