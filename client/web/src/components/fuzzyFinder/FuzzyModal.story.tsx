import { storiesOf } from '@storybook/react'
import React from 'react'
import { MemoryRouter } from 'react-router'

import { CaseInsensitiveFuzzySearch } from '../../fuzzyFinder/CaseInsensitiveFuzzySearch'
import { SearchValue } from '../../fuzzyFinder/FuzzySearch'
import { WebStory } from '../WebStory'

import { Ready } from './FuzzyFinder'
import { FuzzyModal } from './FuzzyModal'

let query = 'readme'
const filenames = ['a.txt', 'b.txt', 'docs/readme.md']
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
    maxResults: 100,
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
