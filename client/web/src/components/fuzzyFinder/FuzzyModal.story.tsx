import { storiesOf } from '@storybook/react'
import React from 'react'
import { MemoryRouter } from 'react-router'

import { CaseInsensitiveFuzzySearch } from '../../fuzzyFinder/CaseInsensitiveFuzzySearch'
import { SearchValue } from '../../fuzzyFinder/FuzzySearch'
import { WebStory } from '../WebStory'

import { Ready } from './FuzzyFinder'
import { FuzzyModal } from './FuzzyModal'

let query = 'client'
const filenames = [
    'babel.config.js',
    'client/README.md',
    'client/branded/.eslintignore',
    'client/branded/.eslintrc.js',
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
    commitID: 'commitID',
    repoName: 'repoName',
    downloadFilenames: () => Promise.resolve(filenames),
    fsm,
    setFsm: () => {},
    focusIndex: 0,
    setFocusIndex: () => {},
    maxResults: 10,
    increaseMaxResults: () => {},
    isVisible: true,
    onClose: () => {},
    query,
    setQuery: (newQuery: string): void => {
        query = newQuery
    },
    caseInsensitiveFileCountThreshold: 100,
}
const { add } = storiesOf('web/FuzzyFinder', module).addDecorator(story => (
    <MemoryRouter initialEntries={[{ pathname: '/' }]}>
        <WebStory>{() => story()}</WebStory>
    </MemoryRouter>
))

add('Ready', () => <FuzzyModal {...defaultProps} />)
