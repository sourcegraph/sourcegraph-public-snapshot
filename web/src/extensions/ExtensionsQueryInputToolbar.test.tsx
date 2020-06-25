import React from 'react'
import { EXTENSION_CATEGORIES } from '../../../shared/src/schema/extensionSchema'
import { extensionsQuery } from './extension/extension'
import { ExtensionsQueryInputToolbar } from './ExtensionsQueryInputToolbar'
import { mount } from 'enzyme'

describe('ExtensionsQueryInputToolbar', () => {
    test('renders', () => {
        expect(mount(<ExtensionsQueryInputToolbar query="q" onQueryChange={() => undefined} />)).toMatchSnapshot()
    })

    test('shows category in query as selected', () => {
        expect(
            mount(
                <ExtensionsQueryInputToolbar
                    query={extensionsQuery({ category: EXTENSION_CATEGORIES[0] })}
                    onQueryChange={() => undefined}
                />
            )
        ).toMatchSnapshot()
    })
})
