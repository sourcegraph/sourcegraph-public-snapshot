jest.mock('../settings/DynamicallyImportedMonacoSettingsEditor', () => ({
    DynamicallyImportedMonacoSettingsEditor: () => 'DynamicallyImportedMonacoSettingsEditor',
}))

import * as H from 'history'
import React from 'react'
import renderer from 'react-test-renderer'
import { noop } from 'rxjs'
import { CodeHostKind } from '../../../shared/src/graphql/schema'
import { SiteAdminCodeHostForm } from './SiteAdminCodeHostForm'

describe('<SiteAdminCodeHostForm />', () => {
    const baseProps = {
        history: H.createMemoryHistory(),
        isLightTheme: true,
        onSubmit: noop,
        onChange: noop,
        jsonSchema: { $id: 'json-schema-id' },
        editorActions: [],
    }

    test('create GitHub', () => {
        const component = renderer.create(
            <SiteAdminCodeHostForm
                {...baseProps}
                input={{
                    kind: CodeHostKind.GITHUB,
                    displayName: 'GitHub',
                    config: '{}',
                }}
                mode="create"
                loading={false}
            />
        )
        expect(component.toJSON()).toMatchSnapshot()
    })
    test('edit GitHub', () => {
        const component = renderer.create(
            <SiteAdminCodeHostForm
                {...baseProps}
                input={{
                    kind: CodeHostKind.GITHUB,
                    displayName: 'GitHub',
                    config: '{}',
                }}
                mode="create"
                loading={false}
            />
        )
        expect(component.toJSON()).toMatchSnapshot()
    })
    test('edit GitHub, loading', () => {
        const component = renderer.create(
            <SiteAdminCodeHostForm
                {...baseProps}
                input={{
                    kind: CodeHostKind.GITHUB,
                    displayName: 'GitHub',
                    config: '{}',
                }}
                mode="create"
                loading={true}
            />
        )
        expect(component.toJSON()).toMatchSnapshot()
    })
})
