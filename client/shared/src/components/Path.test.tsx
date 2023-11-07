import { render } from '@testing-library/react'
import { describe, expect, test } from 'vitest'

import { Path } from './Path'

describe('Path', () => {
    test('no path components', () => {
        expect(render(<Path path="" />).asFragment()).toMatchSnapshot()
    })

    test('1 path component', () => {
        expect(render(<Path path="a" />).asFragment()).toMatchSnapshot()
    })

    test('2 path components', () => {
        expect(render(<Path path="a/b" />).asFragment()).toMatchSnapshot()
    })

    test('3 path components', () => {
        expect(render(<Path path="a/b/c" />).asFragment()).toMatchSnapshot()
    })
})
