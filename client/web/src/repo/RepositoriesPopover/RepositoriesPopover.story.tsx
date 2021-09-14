import { Meta } from '@storybook/react'
import React from 'react'

import { WebStory } from '@sourcegraph/web/src/components/WebStory'

import { RepositoriesPopover } from './RepositoriesPopover'
import { MOCK_REQUESTS } from './RepositoriesPopover.mocks'

const Story: Meta = {
    title: 'web/RepositoriesPopover',

    decorators: [
        story => <WebStory mocks={MOCK_REQUESTS}>{() => <div className="container mt-3">{story()}</div>}</WebStory>,
    ],

    parameters: {
        component: RepositoriesPopover,
    },
}

export default Story

export const RepositoriesPopoverExample = () => <RepositoriesPopover currentRepo="some-repo-id" />
