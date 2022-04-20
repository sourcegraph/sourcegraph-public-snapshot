import { storiesOf } from '@storybook/react'

import { WebStory } from '../components/WebStory'

import { RepositoryMenu, RepositoryMenuContent } from './RepositoryMenu'

const defaultProps = {
    isOpen: true,
    repoName: 'repoName',
    revision: 'commitID',
    filePath: 'foo/bar/baz.bonk',
    settingsCascade: { subjects: null, final: null },
}
const { add } = storiesOf('web/codeintel/RepositoryMenu', module).addDecorator(story => (
    <WebStory>{() => story()}</WebStory>
))

add('Unavailable', () => <RepositoryMenu {...defaultProps} content={RepositoryMenuContent} {...defaultProps} />)
