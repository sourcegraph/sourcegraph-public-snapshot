import React from 'react'
import { mount } from 'enzyme'
import { EditUserProfileForm } from './EditUserProfileForm'

jest.mock('./UserProfileFormFields', () => ({ UserProfileFormFields: 'mock-UserProfileFormFields' }))

describe('EditUserProfileForm', () => {
    test('simple', () =>
        expect(
            mount(
                <EditUserProfileForm
                    authenticatedUser={{ siteAdmin: false }}
                    user={{ id: 'x', viewerCanChangeUsername: true }}
                    initialValue={{ username: 'u', displayName: 'd', avatarURL: 'https://example.com/image.jpg' }}
                    onUpdate={() => {}}
                    after="AFTER"
                />
            )
        ).toMatchSnapshot())
})
