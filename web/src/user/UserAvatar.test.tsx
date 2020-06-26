import React from 'react'
import { UserAvatar } from './UserAvatar'
import { mount } from 'enzyme'

describe('UserAvatar', () => {
    test('no avatar URL', () => expect(mount(<UserAvatar user={{ avatarURL: null }} />).children()).toMatchSnapshot())

    test('with avatar URL', () => expect(mount(<UserAvatar user={{ avatarURL: 'u' }} />).children()).toMatchSnapshot())
})
