import type { Meta, Decorator, StoryFn } from '@storybook/react'
import classNames from 'classnames'

import { WebStory } from '../../../../components/WebStory'
import type { HiddenChangesetApplyPreviewFields } from '../../../../graphql-operations'

import { HiddenChangesetApplyPreviewNode } from './HiddenChangesetApplyPreviewNode'
import { hiddenChangesetApplyPreviewStories } from './storyData'

import styles from './PreviewList.module.scss'

const decorator: Decorator = story => (
    <div className={classNames(styles.previewListGrid, 'p-3 container')}>{story()}</div>
)
const config: Meta = {
    title: 'web/batches/preview/HiddenChangesetApplyPreviewNode',
    decorators: [decorator],
}

export default config

const Template: StoryFn<{ node: HiddenChangesetApplyPreviewFields }> = ({ node }) => (
    <WebStory>{props => <HiddenChangesetApplyPreviewNode {...props} node={node} />}</WebStory>
)

export const ImportChangeset = Template.bind({})
ImportChangeset.args = {
    node: hiddenChangesetApplyPreviewStories['Import changeset'],
}
ImportChangeset.storyName = 'Import changeset'

export const CreateChangeset = Template.bind({})
CreateChangeset.args = {
    node: hiddenChangesetApplyPreviewStories['Create changeset'],
}
CreateChangeset.storyName = 'Create changeset'
