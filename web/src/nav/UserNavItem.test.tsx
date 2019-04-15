import * as H from 'history'
import React from 'react'
import { MemoryRouter } from 'react-router'
import renderer from 'react-test-renderer'
import { setLinkComponent } from '../../../shared/src/components/Link'
import * as GQL from '../../../shared/src/graphql/schema'
import { ThemePreference } from '../theme'
import { UserNavItem } from './UserNavItem'

describe('UserNavItem', () => {
    setLinkComponent((props: any) => <a {...props} />)
    afterAll(() => setLinkComponent(null as any)) // reset global env for other tests

    // tslint:disable-next-line:no-object-literal-type-assertion
    const USER = { username: 'u', url: '/u', settingsURL: '/u/settings' } as GQL.IUser
    const history = H.createMemoryHistory({ keyLength: 0 })

    test('simple', () => {
        expect(
            renderer
                .create(
                    <MemoryRouter>
                        <UserNavItem
                            isLightTheme={true}
                            // tslint:disable-next-line:jsx-no-lambda
                            onThemePreferenceChange={() => void 0}
                            themePreference={ThemePreference.Light}
                            location={history.location}
                            authenticatedUser={USER}
                            showDotComMarketing={true}
                            showDiscussions={true}
                        />
                    </MemoryRouter>
                )
                .toJSON()
        ).toMatchSnapshot()
    })
})
