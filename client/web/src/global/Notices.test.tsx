import { render } from '@testing-library/react'
import React from 'react'

import { Notices } from './Notices'

describe('Notices', () => {
    test('shows notices for location', () =>
        expect(
            render(
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
            render(
                <Notices location="home" settingsCascade={{ subjects: [], final: { notices: null } }} />
            ).asFragment()
        ).toMatchSnapshot())
})
