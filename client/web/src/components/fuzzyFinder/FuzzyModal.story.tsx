import { forceReRender, storiesOf } from '@storybook/react'
import React from 'react'

import { CaseInsensitiveFuzzySearch } from '../../fuzzyFinder/CaseInsensitiveFuzzySearch'
import { SearchValue } from '../../fuzzyFinder/FuzzySearch'
import { WebStory } from '../WebStory'

import { Ready } from './FuzzyFinder'
import { FuzzyModal } from './FuzzyModal'

let query = 'client'
let focusIndex = 0
const INITIAL_MAX_RESULTS = 10
let maxResults = INITIAL_MAX_RESULTS
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
    setFocusIndex: (newFocusIndex: number): void => {
        focusIndex = newFocusIndex
        forceReRender()
    },
    increaseMaxResults: () => {
        maxResults += INITIAL_MAX_RESULTS
        forceReRender()
    },
    isVisible: true,
    onClose: () => {},
    setQuery: (newQuery: string): void => {
        query = newQuery
        forceReRender()
    },
    caseInsensitiveFileCountThreshold: 100,
}
const { add } = storiesOf('web/FuzzyFinder', module).addDecorator(story => <WebStory>{() => story()}</WebStory>)

add('Ready', () => <FuzzyModal {...defaultProps} query={query} focusIndex={focusIndex} maxResults={maxResults} />)
