import React from 'react'
import renderer from 'react-test-renderer'

import { ExtensionRegistrySidenav } from './ExtensionRegistrySidenav'

describe('ExtensionsQueryInputToolbar', () => {
    test('renders', () => {
        expect(
            renderer
                .create(
                    <ExtensionRegistrySidenav
                        activeCategory="Code analysis"
                        onSelectActiveCategory={() => {}}
                        enablementFilter="all"
                        setEnablementFilter={() => {}}
                    />
                )
                .toJSON()
        ).toMatchSnapshot()
    })
})
