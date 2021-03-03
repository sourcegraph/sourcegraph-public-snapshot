import React from 'react'
import { mount } from 'enzyme'
import { EditUserProfileForm } from './EditUserProfileForm'

jest.mock('./UserProfileFormFields', () => ({ UserProfileFormFields: 'mock-UserProfileFormFields' }))

describe('EditUserProfileForm', () => {
    test('simple', () =>
        expect(
            mount(
                <EditUserProfileForm
                    user={{ id: 'x', viewerCanChangeUsername: true }}
                    initialValue={{ username: 'u', displayName: 'd', avatarURL: 'https://example.com/image.jpg' }}
                    after="AFTER"
                />
            )
        ).toMatchSnapshot())
})
