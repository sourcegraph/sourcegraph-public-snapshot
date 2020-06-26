import React from 'react'
import { Timestamp } from './Timestamp'
import { mount } from 'enzyme'

describe('Timestamp', () => {
    test('mocked current time', () => expect(mount(<Timestamp date="2006-01-02" />).children()).toMatchSnapshot())

    test('noAbout', () => expect(mount(<Timestamp date="2006-01-02" noAbout={true} />).children()).toMatchSnapshot())
})
