import { render } from '@testing-library/react'
import React from 'react'

import { ExtensionPermissionModal } from './ExtensionPermissionModal'

describe('ExtensionPermissionModal', () => {
    test('renders', () => {
        expect(
            render(
                <ExtensionPermissionModal
                    extensionID="sourcegraph/typescript"
                    givePermission={() => {}}
                    denyPermission={() => {}}
                />
            ).asFragment()
        ).toMatchSnapshot()
    })
})
