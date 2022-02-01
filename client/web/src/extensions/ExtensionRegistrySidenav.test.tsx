import { render } from '@testing-library/react'
import React from 'react'

import { ExtensionRegistrySidenav } from './ExtensionRegistrySidenav'

describe('ExtensionsQueryInputToolbar', () => {
    test('renders', () => {
        expect(
            render(
                <ExtensionRegistrySidenav
                    selectedCategory="Code analysis"
                    onSelectCategory={() => {}}
                    enablementFilter="all"
                    setEnablementFilter={() => {}}
                    showExperimentalExtensions={true}
                    toggleExperimentalExtensions={() => {}}
                />
            ).asFragment()
        ).toMatchSnapshot()
    })
})
