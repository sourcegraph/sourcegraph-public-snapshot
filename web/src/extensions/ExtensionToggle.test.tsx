import React from 'react'
import { ConfiguredRegistryExtension } from '../../../shared/src/extensions/extension'
import { PlatformContext } from '../../../shared/src/platform/context'
import { ConfiguredSubjectOrError, SettingsSubject } from '../../../shared/src/settings/settings'
import { ExtensionToggle } from './ExtensionToggle'
import { mount } from 'enzyme'

describe('ExtensionToggle', () => {
    const NOOP_PLATFORM_CONTEXT: PlatformContext = {} as any
    const SUBJECT: ConfiguredSubjectOrError = {
        lastID: null,
        settings: {},
        // eslint-disable-next-line @typescript-eslint/consistent-type-assertions
        subject: { __typename: 'User', id: 'u', viewerCanAdminister: true } as SettingsSubject,
    }
    const EXTENSION: Pick<ConfiguredRegistryExtension, 'id'> = {
        id: 'x/y',
    }

    test('extension not present in settings', () => {
        expect(
            mount(
                <ExtensionToggle
                    extension={EXTENSION}
                    settingsCascade={{ final: { extensions: {} }, subjects: [SUBJECT] }}
                    platformContext={NOOP_PLATFORM_CONTEXT}
                />
            ).children()
        ).toMatchSnapshot()
    })

    test('extension enabled in settings', () => {
        expect(
            mount(
                <ExtensionToggle
                    extension={EXTENSION}
                    settingsCascade={{ final: { extensions: { 'x/y': true } }, subjects: [SUBJECT] }}
                    platformContext={NOOP_PLATFORM_CONTEXT}
                />
            ).children()
        ).toMatchSnapshot()
    })

    test('extension disabled in settings', () => {
        expect(
            mount(
                <ExtensionToggle
                    extension={EXTENSION}
                    settingsCascade={{ final: { extensions: { 'x/y': false } }, subjects: [SUBJECT] }}
                    platformContext={NOOP_PLATFORM_CONTEXT}
                />
            ).children()
        ).toMatchSnapshot()
    })
})
