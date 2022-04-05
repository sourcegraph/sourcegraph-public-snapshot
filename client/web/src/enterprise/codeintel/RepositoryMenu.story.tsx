import React from 'react'

import { storiesOf } from '@storybook/react'

import { WebStory } from '../../components/WebStory'

import { RepositoryMenuContent } from './RepositoryMenu'
import { UseCodeIntelStatusParameters, UseCodeIntelStatusResult } from './useCodeIntelStatus'

const defaultProps = {
    repoName: 'repoName',
    revision: 'commitID',
    filePath: 'foo/bar/baz.bonk',
    settingsCascade: { subjects: null, final: null },
    useCodeIntelStatus: (parameters: UseCodeIntelStatusParameters): UseCodeIntelStatusResult => ({
        data: {},
        loading: false,
    }),
}
const { add } = storiesOf('web/codeintel/enterprise/RepositoryMenu', module).addDecorator(story => (
    <WebStory>{() => story()}</WebStory>
))

add('Basic', () => <RepositoryMenuContent {...defaultProps} />)
