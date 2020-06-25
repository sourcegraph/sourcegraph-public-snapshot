import React from 'react'
import { PlatformContext } from '../../../shared/src/platform/context'
import { ExtensionCard } from './ExtensionCard'
import { mount } from 'enzyme'

describe('ExtensionCard', () => {
    const NOOP_PLATFORM_CONTEXT: PlatformContext = {} as any

    test('renders', () => {
        expect(
            mount(
                <ExtensionCard
                    node={{
                        id: 'x/y',
                        manifest: {
                            activationEvents: ['*'],
                            description: 'd',
                            url: 'https://example.com',
                            icon: 'data:image/png,abcd',
                        },
                        registryExtension: {
                            id: 'abcd1234',
                            extensionIDWithoutRegistry: 'x/y',
                            url: 'extensions/x/y',
                            isWorkInProgress: false,
                            viewerCanAdminister: false,
                        },
                    }}
                    subject={{ id: 'u', viewerCanAdminister: false }}
                    settingsCascade={{ final: null, subjects: null }}
                    platformContext={NOOP_PLATFORM_CONTEXT}
                />
            )
        ).toMatchSnapshot()
    })
})
