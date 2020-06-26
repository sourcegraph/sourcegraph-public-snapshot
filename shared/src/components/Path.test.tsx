import React from 'react'
import { Path } from './Path'
import { mount } from 'enzyme'

describe('Path', () => {
    test('no path components', () => {
        expect(mount(<Path path="" />).children()).toMatchSnapshot()
    })

    test('1 path component', () => {
        expect(mount(<Path path="a" />).children()).toMatchSnapshot()
    })

    test('2 path components', () => {
        expect(mount(<Path path="a/b" />).children()).toMatchSnapshot()
    })

    test('3 path components', () => {
        expect(mount(<Path path="a/b/c" />).children()).toMatchSnapshot()
    })
})
