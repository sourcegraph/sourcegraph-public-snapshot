import { describe, expect, it } from '@jest/globals'
import { render } from '@testing-library/react'

import { H5 } from './H5'

describe('H5', () => {
    it('should render correctly', () => {
        expect(
            render(
                <H5 alignment="left" mode="single-line">
                    This is H5
                </H5>
            ).asFragment()
        ).toMatchSnapshot()
    })

    it('should render correctly with `as`', () => {
        expect(
            render(
                <H5 alignment="left" mode="single-line" as="p">
                    This is a p
                </H5>
            ).asFragment()
        ).toMatchSnapshot()
    })
})
