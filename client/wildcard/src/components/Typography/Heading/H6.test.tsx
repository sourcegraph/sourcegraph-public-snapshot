import { render } from '@testing-library/react'
import { describe, expect, it } from 'vitest'

import { H6 } from './H6'

describe('H6', () => {
    it('should render correctly', () => {
        expect(
            render(
                <H6 alignment="left" mode="single-line">
                    This is H6
                </H6>
            ).asFragment()
        ).toMatchSnapshot()
    })

    it('should render correctly with `as`', () => {
        expect(
            render(
                <H6 alignment="left" mode="single-line" as="p">
                    This is a p
                </H6>
            ).asFragment()
        ).toMatchSnapshot()
    })
})
