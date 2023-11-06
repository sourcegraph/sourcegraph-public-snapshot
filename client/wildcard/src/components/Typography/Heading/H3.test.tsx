import { describe, expect, it } from '@jest/globals'
import { render } from '@testing-library/react'

import { H3 } from '../..'

describe('H3', () => {
    it('should render correctly', () => {
        expect(
            render(
                <H3 alignment="left" mode="single-line">
                    This is H3
                </H3>
            ).asFragment()
        ).toMatchSnapshot()
    })

    it('should render correctly with `as`', () => {
        expect(
            render(
                <H3 alignment="left" mode="single-line" as="p">
                    This is a p
                </H3>
            ).asFragment()
        ).toMatchSnapshot()
    })
})
