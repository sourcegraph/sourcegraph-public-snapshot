import { render } from '@testing-library/react'

import { ConfiguredRegistryExtension } from '@sourcegraph/shared/src/extensions/extension'
import { PlatformContext } from '@sourcegraph/shared/src/platform/context'
import { ConfiguredSubjectOrError, SettingsSubject } from '@sourcegraph/shared/src/settings/settings'

import { ExtensionToggle } from './ExtensionToggle'

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
            render(
                <ExtensionToggle
                    extensionID={EXTENSION.id}
                    enabled={false}
                    settingsCascade={{ final: { extensions: {} }, subjects: [SUBJECT] }}
                    platformContext={NOOP_PLATFORM_CONTEXT}
                    subject={SUBJECT.subject}
                />
            ).asFragment()
        ).toMatchSnapshot()
    })

    test('extension enabled in settings', () => {
        expect(
            render(
                <ExtensionToggle
                    extensionID={EXTENSION.id}
                    enabled={true}
                    settingsCascade={{ final: { extensions: { 'x/y': true } }, subjects: [SUBJECT] }}
                    platformContext={NOOP_PLATFORM_CONTEXT}
                    subject={SUBJECT.subject}
                />
            ).asFragment()
        ).toMatchSnapshot()
    })

    test('extension disabled in settings', () => {
        expect(
            render(
                <ExtensionToggle
                    extensionID={EXTENSION.id}
                    enabled={false}
                    settingsCascade={{ final: { extensions: { 'x/y': false } }, subjects: [SUBJECT] }}
                    platformContext={NOOP_PLATFORM_CONTEXT}
                    subject={SUBJECT.subject}
                />
            ).asFragment()
        ).toMatchSnapshot()
    })
})
