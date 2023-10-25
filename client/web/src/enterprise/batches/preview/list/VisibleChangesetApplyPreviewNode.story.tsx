import type { Meta, Decorator, StoryFn, StoryObj } from '@storybook/react'
import classNames from 'classnames'
import { of } from 'rxjs'

import { WebStory } from '../../../../components/WebStory'
import type { VisibleChangesetApplyPreviewFields } from '../../../../graphql-operations'

import { visibleChangesetApplyPreviewNodeStories } from './storyData'
import { VisibleChangesetApplyPreviewNode } from './VisibleChangesetApplyPreviewNode'

import styles from './PreviewList.module.scss'

const decorator: Decorator = story => (
    <div className={classNames(styles.previewListGrid, 'p-3 container')}>{story()}</div>
)

const config: Meta<typeof VisibleChangesetApplyPreviewNode> = {
    title: 'web/batches/preview/VisibleChangesetApplyPreviewNode',
    decorators: [decorator],
}

export default config

const queryEmptyFileDiffs = () => of({ totalCount: 0, pageInfo: { endCursor: null, hasNextPage: false }, nodes: [] })

const stories = visibleChangesetApplyPreviewNodeStories(true)

const Template: StoryFn<{ node: VisibleChangesetApplyPreviewFields }> = ({ node }) => (
    <WebStory>
        {props => (
            <VisibleChangesetApplyPreviewNode
                {...props}
                node={node}
                authenticatedUser={{
                    url: '/users/alice',
                    displayName: 'Alice',
                    username: 'alice',
                    emails: [{ email: 'alice@email.test', isPrimary: true, verified: true }],
                }}
                queryChangesetSpecFileDiffs={queryEmptyFileDiffs}
            />
        )}
    </WebStory>
)

type Story = StoryObj<typeof config>

export const ImportChangeset: Story = Template.bind({})
ImportChangeset.args = { node: stories['Import changeset'] }
ImportChangeset.storyName = 'Import changeset'

export const CreateChangesetPublished: Story = Template.bind({})
CreateChangesetPublished.args = { node: stories['Create changeset published'] }
CreateChangesetPublished.storyName = 'Create changeset published'

export const CreateChangesetDraft: Story = Template.bind({})
CreateChangesetDraft.args = { node: stories['Create changeset draft'] }
CreateChangesetDraft.storyName = 'Create changeset draft'

export const CreateChangesetNotPublished: Story = Template.bind({})
CreateChangesetNotPublished.args = { node: stories['Create changeset not published'] }
CreateChangesetNotPublished.storyName = 'Create changeset not published'

export const UpdateChangesetTitle: Story = Template.bind({})
UpdateChangesetTitle.args = { node: stories['Update changeset title'] }
UpdateChangesetTitle.storyName = 'Update changeset title'

export const UpdateChangesetBody: Story = Template.bind({})
UpdateChangesetBody.args = { node: stories['Update changeset body'] }
UpdateChangesetBody.storyName = 'Update changeset body'

export const UndraftChangeset: Story = Template.bind({})
UndraftChangeset.args = { node: stories['Undraft changeset'] }
UndraftChangeset.storyName = 'Undraft changeset'

export const ReopenChangeset: Story = Template.bind({})
ReopenChangeset.args = { node: stories['Reopen changeset'] }
ReopenChangeset.storyName = 'Reopen changeset'

export const ChangeDiff: Story = Template.bind({})
ChangeDiff.args = { node: stories['Change diff'] }
ChangeDiff.storyName = 'Change diff'

export const CloseChangeset: Story = Template.bind({})
CloseChangeset.args = { node: stories['Close changeset'] }
CloseChangeset.storyName = 'Close changeset'

export const DetachChangeset: Story = Template.bind({})
DetachChangeset.args = { node: stories['Detach changeset'] }
DetachChangeset.storyName = 'Detach changeset'

export const ChangeBaseRef: Story = Template.bind({})
ChangeBaseRef.args = { node: stories['Change base ref'] }
ChangeBaseRef.storyName = 'Change base ref'

export const UpdateCommitMessage: Story = Template.bind({})
UpdateCommitMessage.args = { node: stories['Update commit message'] }
UpdateCommitMessage.storyName = 'Update commit message'

export const UpdateCommitAuthor: Story = Template.bind({})
UpdateCommitAuthor.args = { node: stories['Update commit author'] }
UpdateCommitAuthor.storyName = 'Update commit author'

export const ForkedRepo: Story = Template.bind({})
ForkedRepo.args = { node: stories['Forked repo'] }
ForkedRepo.storyName = 'Forked repo'

export const ReattachChangeset: Story = Template.bind({})
ReattachChangeset.args = { node: stories['Reattach changeset'] }
ReattachChangeset.storyName = 'Reattach Changeset'
