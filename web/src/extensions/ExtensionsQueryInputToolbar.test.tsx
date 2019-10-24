import React from 'react'
import renderer from 'react-test-renderer'
import { EXTENSION_CATEGORIES } from '../../../shared/src/schema/extensionSchema'
import { extensionsQuery } from './extension/extension'
import { ExtensionsQueryInputToolbar } from './ExtensionsQueryInputToolbar'

describe('ExtensionsQueryInputToolbar', () => {
    test('renders', () => {
        expect(
            renderer.create(<ExtensionsQueryInputToolbar query="q" onQueryChange={() => undefined} />).toJSON()
        ).toMatchSnapshot()
    })

    test('shows category in query as selected', () => {
        expect(
            renderer
                .create(
                    <ExtensionsQueryInputToolbar
                        query={extensionsQuery({ category: EXTENSION_CATEGORIES[0] })}
                        onQueryChange={() => undefined}
                    />
                )
                .toJSON()
        ).toMatchSnapshot()
    })
})
