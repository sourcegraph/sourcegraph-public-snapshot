import React from 'react'
import { Notices } from './Notices'
import * as H from 'history'
import { mount } from 'enzyme'

describe('Notices', () => {
    test('shows notices for location', () =>
        expect(
            mount(
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
                    history={H.createMemoryHistory()}
                />
            ).children()
        ).toMatchSnapshot())

    test('no notices', () =>
        expect(
            mount(
                <Notices
                    location="home"
                    history={H.createMemoryHistory()}
                    settingsCascade={{ subjects: [], final: { notices: null } }}
                />
            ).children()
        ).toMatchSnapshot())
})
