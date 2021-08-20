import React from 'react'
import renderer from 'react-test-renderer'

import { ExtensionRegistrySidenav } from './ExtensionRegistrySidenav'

describe('ExtensionsQueryInputToolbar', () => {
    test('renders', () => {
        expect(
            renderer
                .create(
                    <ExtensionRegistrySidenav
                        selectedCategory="Code analysis"
                        onSelectCategory={() => {}}
                        enablementFilter="all"
                        setEnablementFilter={() => {}}
                        showExperimentalExtensions={true}
                        toggleExperimentalExtensions={() => {}}
                    />
                )
                .toJSON()
        ).toMatchSnapshot()
    })
})
