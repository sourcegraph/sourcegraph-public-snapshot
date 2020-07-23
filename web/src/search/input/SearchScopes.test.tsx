import { createMemoryHistory } from 'history'
import renderer from 'react-test-renderer'
import React from 'react'
import { MemoryRouter } from 'react-router'
import { SearchScopes } from './SearchScopes'
import * as GQL from '../../../../shared/src/graphql/schema'
import { Settings } from '../../schema/settings.schema'

const BASE_PROPS = {
    authenticatedUser: {},
    history: createMemoryHistory(),
    query: 'abc',
    patternType: GQL.SearchPatternType.literal,
    versionContext: undefined,
}

describe('SearchScopes', () => {
    test('empty', () =>
        expect(
            renderer
                .create(
                    <MemoryRouter>
                        <SearchScopes {...BASE_PROPS} settingsCascade={{ final: {}, subjects: [] }} />
                    </MemoryRouter>
                )
                .toJSON()
        ).toMatchSnapshot())

    test('with scopes', () => {
        const settings: Settings = { 'search.scopes': [{ name: 'n', value: 'v', description: 'd', id: 'i' }] }
        expect(
            renderer
                .create(
                    <MemoryRouter>
                        <SearchScopes {...BASE_PROPS} settingsCascade={{ final: settings, subjects: [] }} />
                    </MemoryRouter>
                )
                .toJSON()
        ).toMatchSnapshot()
    })
})
