import React from 'react'

import { H1 } from '@sourcegraph/wildcard'

import { CodyIcon } from '../../cody/CodyIcon'

import { CodyChat } from './CodyChat'

/**
 * For Sourcegraph team members only. For instructions, see
 * https://docs.google.com/document/d/1u7HYPmJFtDANtBgczzmAR0BmhM86drwDXCqx-F2jTEE/edit#.
 */
export const CodyPage: React.FunctionComponent<{}> = () => (
    <div>
        <H1>
            <CodyIcon /> Cody
        </H1>
        <CodyChat />
    </div>
)
