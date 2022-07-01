import { Meta, DecoratorFn, Story } from '@storybook/react'
import classNames from 'classnames'

import { WebStory } from '../../../../components/WebStory'
import { ChangesetSpecType, HiddenChangesetApplyPreviewFields } from '../../../../graphql-operations'

import { HiddenChangesetApplyPreviewNode } from './HiddenChangesetApplyPreviewNode'

import styles from './PreviewList.module.scss'

const decorator: DecoratorFn = story => (
    <div className={classNames(styles.previewListGrid, 'p-3 container')}>{story()}</div>
)
const config: Meta = {
    title: 'web/batches/preview/HiddenChangesetApplyPreviewNode',
    decorators: [decorator],
}

export default config

const hiddenChangesetApplyPreviewStories: Record<string, HiddenChangesetApplyPreviewFields> = {
    importChangeset: {
        __typename: 'HiddenChangesetApplyPreview',
        targets: {
            __typename: 'HiddenApplyPreviewTargetsAttach',
            changesetSpec: {
                __typename: 'HiddenChangesetSpec',
                id: 'someidh1',
                type: ChangesetSpecType.EXISTING,
            },
        },
    },
    createChangeset: {
        __typename: 'HiddenChangesetApplyPreview',
        targets: {
            __typename: 'HiddenApplyPreviewTargetsAttach',
            changesetSpec: {
                __typename: 'HiddenChangesetSpec',
                id: 'someidh2',
                type: ChangesetSpecType.BRANCH,
            },
        },
    },
}

export const ImportChangeset: Story = () => (
    <WebStory>
        {props => (
            <HiddenChangesetApplyPreviewNode {...props} node={hiddenChangesetApplyPreviewStories.importChangeset} />
        )}
    </WebStory>
)

ImportChangeset.storyName = 'Import changeset'

export const CreateChangeset: Story = () => (
    <WebStory>
        {props => (
            <HiddenChangesetApplyPreviewNode {...props} node={hiddenChangesetApplyPreviewStories.createChangeset} />
        )}
    </WebStory>
)

CreateChangeset.storyName = 'Create changeset'
