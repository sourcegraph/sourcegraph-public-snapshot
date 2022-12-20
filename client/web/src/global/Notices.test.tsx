import { renderWithBrandedContext } from '@sourcegraph/wildcard'

import { Notices } from './Notices'

describe('Notices', () => {
    test('shows notices for location', () =>
        expect(
            renderWithBrandedContext(
                <Notices
                    location="home"
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
                />
            ).asFragment()
        ).toMatchSnapshot())

    test('no notices', () =>
        expect(
            renderWithBrandedContext(
                <Notices location="home" settingsCascade={{ subjects: [], final: { notices: undefined } }} />
            ).asFragment()
        ).toMatchSnapshot())
})
