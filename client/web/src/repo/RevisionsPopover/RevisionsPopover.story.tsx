import { ComponentMeta, ComponentStory } from '@storybook/react'
import React from 'react'

import { WebStory } from '../../components/WebStory'

import { RevisionsPopover } from './RevisionsPopover'
import { revisionPopoverUserActions } from './RevisionsPopover.actions'
import { MOCK_PROPS, MOCK_REQUESTS } from './RevisionsPopover.mocks'

const config: ComponentMeta<typeof RevisionsPopover> = {
    /* 👇 The title prop is optional.
     * See https://storybook.js.org/docs/react/configure/overview#configure-story-loading
     * to learn how to generate automatic titles
     */
    title: 'web/RevisionsPopover',
    component: RevisionsPopover,
}

export default config

const Template: ComponentStory<typeof RevisionsPopover> = templateProps => (
    <WebStory
        mocks={MOCK_REQUESTS}
        initialEntries={[{ pathname: `/${MOCK_PROPS.repoName}` }]}
        // Can't utilise loose mocking here as the commit/branch requests use the same operations just with different variables
        useStrictMocking={true}
    >
        {webStoryProps => <RevisionsPopover {...webStoryProps} {...templateProps} {...MOCK_PROPS} />}
    </WebStory>
)

export const RevisionsPopoverBranches = Template.bind({})
RevisionsPopoverBranches.play = async ({ canvasElement }) => {
    const { selectBranchTab, focusGitReference } = revisionPopoverUserActions(canvasElement)
    await selectBranchTab()
    focusGitReference()
}

export const RevisionsPopoverTags = Template.bind({})
RevisionsPopoverTags.play = async ({ canvasElement }) => {
    const { selectTagTab, focusGitReference } = revisionPopoverUserActions(canvasElement)
    await selectTagTab()
    focusGitReference()
}

export const RevisionsPopoverCommits = Template.bind({})
RevisionsPopoverCommits.play = async ({ canvasElement }) => {
    const { selectCommitTab } = revisionPopoverUserActions(canvasElement)
    await selectCommitTab()
}
