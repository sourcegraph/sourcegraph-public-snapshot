import { describe, expect, test } from '@jest/globals'

import { SettingsProvider } from '@sourcegraph/shared/src/settings/settings'
import { renderWithBrandedContext } from '@sourcegraph/wildcard/src/testing'

import { Notices } from './Notices'

describe('Notices', () => {
    test('shows notices for location', () =>
        expect(
            renderWithBrandedContext(
                <SettingsProvider
                    settingsCascade={{
                        subjects: [],
                        final: {
                            notices: [
                                { message: 'a', location: 'home' },
                                { message: 'a', location: 'home', dismissible: true },
                                { message: 'b', location: 'top' },
                            ],
                        },
                    }}
                >
                    <Notices location="home" />
                </SettingsProvider>
            ).asFragment()
        ).toMatchSnapshot())

    test('no notices', () =>
        expect(
            renderWithBrandedContext(
                <SettingsProvider settingsCascade={{ subjects: [], final: { notices: undefined } }}>
                    <Notices location="home" />
                </SettingsProvider>
            ).asFragment()
        ).toMatchSnapshot())
})
