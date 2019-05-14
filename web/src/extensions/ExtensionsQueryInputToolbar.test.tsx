import React from 'react'
import renderer from 'react-test-renderer'
import { EXTENSION_CATEGORIES } from '../../../shared/src/schema/extensionSchema'
import { extensionsQuery } from './extension/extension'
import { ExtensionsQueryInputToolbar } from './ExtensionsQueryInputToolbar'

describe('ExtensionsQueryInputToolbar', () => {
    test('renders', () => {
        expect(
            // tslint:disable-next-line:jsx-no-lambda
            renderer.create(<ExtensionsQueryInputToolbar query="q" onQueryChange={() => void 0} />).toJSON()
        ).toMatchSnapshot()
    })

    test('shows category in query as selected', () => {
        expect(
            renderer
                .create(
                    <ExtensionsQueryInputToolbar
                        query={extensionsQuery({ category: EXTENSION_CATEGORIES[0] })}
                        // tslint:disable-next-line:jsx-no-lambda
                        onQueryChange={() => void 0}
                    />
                )
                .toJSON()
        ).toMatchSnapshot()
    })
})
