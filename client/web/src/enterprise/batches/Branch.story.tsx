import type { Meta, Story } from '@storybook/react'

import { WebStory } from '../../components/WebStory'

import { BranchMerge } from './Branch'

const config: Meta = {
    title: 'web/batches/Branch',
}

export default config

export const Forked: Story = () => (
    <WebStory>
        {() => <BranchMerge baseRef="main" forkTarget={{ pushUser: false, namespace: 'org' }} headRef="branch" />}
    </WebStory>
)

export const WillBeForkedIntoTheUser: Story = () => (
    <WebStory>
        {() => <BranchMerge baseRef="main" forkTarget={{ pushUser: true, namespace: 'org' }} headRef="branch" />}
    </WebStory>
)

WillBeForkedIntoTheUser.storyName = 'Will be forked into the user'

export const Unforked: Story = () => <WebStory>{() => <BranchMerge baseRef="main" headRef="branch" />}</WebStory>
