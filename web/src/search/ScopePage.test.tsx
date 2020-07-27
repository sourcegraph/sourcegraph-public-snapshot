import React from 'react'
import { MemoryRouter } from 'react-router'
import renderer from 'react-test-renderer'
import { ScopePage } from './ScopePage'
import * as H from 'history'
import { SearchPatternType } from '../../../shared/src/graphql/schema'

jest.mock('./input/QueryInput', () => ({ QueryInput: 'QueryInput' }))
jest.mock('./input/SearchButton', () => ({ SearchButton: 'SearchButton' }))

const history = H.createMemoryHistory()
const location = H.createLocation('/')

describe('ScopePage', () => {
    test('renders', () =>
        expect(
            renderer
                .create(
                    <MemoryRouter>
                        <ScopePage
                            match={{ params: { id: 'my-scope-id' }, isExact: true, path: '/', url: '/' }}
                            settingsCascade={{
                                final: {
                                    'search.scopes': [
                                        { id: 'my-scope-id', description: 'my description', value: 'my-scope-value' },
                                    ],
                                },
                                subjects: [],
                            }}
                            authenticatedUser={null}
                            caseSensitive={false}
                            onNavbarQueryChange={() => undefined}
                            patternType={SearchPatternType.literal}
                            setCaseSensitivity={() => undefined}
                            setPatternType={() => undefined}
                            history={history}
                            location={location}
                            copyQueryButton={false}
                            versionContext={undefined}
                            globbing={false}
                        />
                    </MemoryRouter>
                )
                .toJSON()
        ).toMatchSnapshot())
})
