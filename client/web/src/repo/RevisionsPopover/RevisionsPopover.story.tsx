import { Meta } from '@storybook/react'
import React, { useMemo } from 'react'

import { WebStory } from '@sourcegraph/web/src/components/WebStory'

import { RevisionsPopover } from './RevisionsPopover'
import { MOCK_PROPS, MOCK_REQUESTS } from './RevisionsPopover.mocks'

import { LAST_TAB_STORAGE_KEY } from '.'

const Story: Meta = {
    title: 'web/RevisionsPopover',

    decorators: [
        story => (
            <WebStory
                mocks={MOCK_REQUESTS}
                initialEntries={[{ pathname: `/${MOCK_PROPS.repoName}` }]}
                // Can't utilise loose mocking here as the commit/branch requests use the same operations just with different variables
                useStrictMocking={true}
            >
                {() => <div className="container mt-3">{story()}</div>}
            </WebStory>
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

interface RevisionsPopoverStoryProps {
    initialTabIndex: number
}

const RevisionsPopoverStory: React.FunctionComponent<RevisionsPopoverStoryProps> = ({ initialTabIndex }) => {
    // Ensure we have prepared the correct tab index before we render
    useMemo(() => {
        localStorage.setItem(LAST_TAB_STORAGE_KEY, initialTabIndex.toString())
    }, [initialTabIndex])

    return <RevisionsPopover {...MOCK_PROPS} />
}

export const RevisionsPopoverBranches = () => <RevisionsPopoverStory initialTabIndex={0} />
export const RevisionsPopoverTags = () => <RevisionsPopoverStory initialTabIndex={1} />
export const RevisionsPopoverCommits = () => <RevisionsPopoverStory initialTabIndex={2} />
