import { describe, expect, test } from '@jest/globals'
import { render } from '@testing-library/react'

import { UserAvatar } from './UserAvatar'

describe('UserAvatar', () => {
    test('no avatar URL', () =>
        expect(
            render(<UserAvatar user={{ avatarURL: null, username: 'test', displayName: 't' }} />).asFragment()
        ).toMatchSnapshot())

    test('with avatar URL', () =>
        expect(
            render(<UserAvatar user={{ avatarURL: 'u', username: 'test', displayName: 't' }} />).asFragment()
        ).toMatchSnapshot())
})
