import { storiesOf } from '@storybook/react'
import classNames from 'classnames'
import React from 'react'

import { WebStory } from '../../../../components/WebStory'
import { ChangesetSpecType, HiddenChangesetApplyPreviewFields } from '../../../../graphql-operations'

import { HiddenChangesetApplyPreviewNode } from './HiddenChangesetApplyPreviewNode'
import styles from './PreviewList.module.scss'

const { add } = storiesOf('web/batches/preview/HiddenChangesetApplyPreviewNode', module).addDecorator(story => (
    <div className={classNames(styles.previewListGrid, 'p-3 container')}>{story()}</div>
))

export const hiddenChangesetApplyPreviewStories: Record<string, HiddenChangesetApplyPreviewFields> = {
    'Import changeset': {
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
    'Create changeset': {
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

for (const storyName of Object.keys(hiddenChangesetApplyPreviewStories)) {
    add(storyName, () => (
        <WebStory>
            {props => (
                <HiddenChangesetApplyPreviewNode {...props} node={hiddenChangesetApplyPreviewStories[storyName]} />
            )}
        </WebStory>
    ))
}
