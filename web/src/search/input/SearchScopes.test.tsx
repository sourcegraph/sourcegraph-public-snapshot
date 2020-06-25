import { createMemoryHistory } from 'history'
import React from 'react'
import { MemoryRouter } from 'react-router'
import { SearchScopes } from './SearchScopes'
import * as GQL from '../../../../shared/src/graphql/schema'
import { Settings } from '../../schema/settings.schema'
import { mount } from 'enzyme'

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
            mount(
                <MemoryRouter>
                    <SearchScopes {...BASE_PROPS} settingsCascade={{ final: {}, subjects: [] }} />
                </MemoryRouter>
            )
        ).toMatchSnapshot())

    test('with scopes', () => {
        const settings: Settings = { 'search.scopes': [{ name: 'n', value: 'v', description: 'd', id: 'i' }] }
        expect(
            mount(
                <MemoryRouter>
                    <SearchScopes {...BASE_PROPS} settingsCascade={{ final: settings, subjects: [] }} />
                </MemoryRouter>
            )
        ).toMatchSnapshot()
    })
})
