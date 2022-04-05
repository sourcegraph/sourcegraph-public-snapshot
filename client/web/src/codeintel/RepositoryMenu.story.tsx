import React from 'react'

import { storiesOf } from '@storybook/react'

import { WebStory } from '../components/WebStory'

import { RepositoryMenuContent } from './RepositoryMenu'

const defaultProps = {
    repoName: 'repoName',
    revision: 'commitID',
    filePath: 'foo/bar/baz.bonk',
    settingsCascade: { subjects: null, final: null },
}
const { add } = storiesOf('web/codeintel/RepositoryMenu', module).addDecorator(story => (
    <WebStory>{() => story()}</WebStory>
))

add('Basic', () => <RepositoryMenuContent {...defaultProps} />)
