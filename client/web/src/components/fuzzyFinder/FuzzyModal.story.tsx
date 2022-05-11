import { storiesOf } from '@storybook/react'

import { CaseInsensitiveFuzzySearch } from '../../fuzzyFinder/CaseInsensitiveFuzzySearch'
import { SearchValue } from '../../fuzzyFinder/FuzzySearch'
import { WebStory } from '../WebStory'

import { Ready } from './FuzzyFinder'
import { FuzzyModal } from './FuzzyModal'

const filenames = [
    'babel.config.js',
    'client/README.md',
    'client/branded/.eslintignore',
    'client/branded/.eslintrc.js',
    // This line is intentionally long to test what happens during horizontal overflows
    'client/branded/.src/components/BrandedStory.tsx/client/branded/srcndedStory.tsx/client/branded/src/components/BrandedStory.tsx/client/branded/src/components/BrandedStory.tsx',
    'client/branded/.stylelintrc.json',
    'client/branded/README.md',
    'client/branded/babel.config.js',
    'client/branded/jest.config.js',
    'client/branded/package.json',
    'client/branded/src/components/CodeSnippet.tsx',
    'client/branded/src/components/Form.tsx',
    'client/branded/src/components/LoaderInput.scss',
    'client/branded/src/components/LoaderInput.story.tsx',
]
const searchValues: SearchValue[] = filenames.map(filename => ({ text: filename }))
const fuzzy = new CaseInsensitiveFuzzySearch(searchValues)
const fsm: Ready = { key: 'ready', fuzzy }
const defaultProps = {
    repoName: 'repoName',
    commitID: 'commitID',
    initialMaxResults: 10,
    initialQuery: 'clientb',
    downloadFilenames: filenames,
    isLoading: false,
    isError: undefined,
    onClose: () => {},
    fsm,
    setFsm: () => {},
}
const { add } = storiesOf('web/FuzzyFinder', module).addDecorator(story => <WebStory>{() => story()}</WebStory>)

add('Ready', () => <FuzzyModal {...defaultProps} />)
