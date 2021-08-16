import { Meta } from '@storybook/react'
import React from 'react'

import { BrandedStory } from '@sourcegraph/branded/src/components/BrandedStory'
import webStyles from '@sourcegraph/web/src/SourcegraphWebApp.scss'

import { RevisionsPopover } from './RevisionsPopover'
import { MOCK_PROPS, MOCK_REQUESTS } from './RevisionsPopover.mocks'

const Story: Meta = {
    title: 'web/RevisionsPopover',

    decorators: [
        story => (
            <BrandedStory
                styles={webStyles}
                mocks={MOCK_REQUESTS}
                initialEntries={[{ pathname: `/${MOCK_PROPS.repoName}` }]}
                // Can't utilise loose mocking here as the commit/branch requests use the same operations just with different variables
                useStrictMocking={true}
            >
                {() => <div className="container mt-3">{story()}</div>}
            </BrandedStory>
        ),
    ],

    parameters: {
        component: RevisionsPopover,
        design: {
            type: 'figma',
            name: 'Figma',
            url:
                'https://www.figma.com/file/NIsN34NH7lPu04olBzddTw/Design-Refresh-Systemization-source-of-truth?node-id=954%3A2161',
        },
    },
}

export default Story

export const RevisionsPopoverExample = () => <RevisionsPopover {...MOCK_PROPS} />
