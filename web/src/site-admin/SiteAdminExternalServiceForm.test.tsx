jest.mock('../settings/DynamicallyImportedMonacoSettingsEditor', () => ({
    DynamicallyImportedMonacoSettingsEditor: () => 'DynamicallyImportedMonacoSettingsEditor',
}))

import * as H from 'history'
import React from 'react'
import { noop } from 'rxjs'
import { ExternalServiceKind } from '../../../shared/src/graphql/schema'
import { SiteAdminExternalServiceForm } from './SiteAdminExternalServiceForm'
import { mount } from 'enzyme'

describe('<SiteAdminExternalServiceForm />', () => {
    const baseProps = {
        history: H.createMemoryHistory(),
        isLightTheme: true,
        onSubmit: noop,
        onChange: noop,
        jsonSchema: { $id: 'json-schema-id' },
        editorActions: [],
    }

    test('create GitHub', () => {
        expect(
            mount(
                <SiteAdminExternalServiceForm
                    {...baseProps}
                    input={{
                        kind: ExternalServiceKind.GITHUB,
                        displayName: 'GitHub',
                        config: '{}',
                    }}
                    mode="create"
                    loading={false}
                />
            )
        ).toMatchSnapshot()
    })
    test('edit GitHub', () => {
        expect(
            mount(
                <SiteAdminExternalServiceForm
                    {...baseProps}
                    input={{
                        kind: ExternalServiceKind.GITHUB,
                        displayName: 'GitHub',
                        config: '{}',
                    }}
                    mode="create"
                    loading={false}
                />
            )
        ).toMatchSnapshot()
    })
    test('edit GitHub, loading', () => {
        expect(
            mount(
                <SiteAdminExternalServiceForm
                    {...baseProps}
                    input={{
                        kind: ExternalServiceKind.GITHUB,
                        displayName: 'GitHub',
                        config: '{}',
                    }}
                    mode="create"
                    loading={true}
                />
            )
        ).toMatchSnapshot()
    })
})
