import { describe, expect, it } from '@jest/globals'
import { render } from '@testing-library/react'

import { H2 } from './H2'

describe('H2', () => {
    it('should render correctly', () => {
        expect(
            render(
                <H2 alignment="left" mode="single-line">
                    This is H2
                </H2>
            ).asFragment()
        ).toMatchSnapshot()
    })

    it('renders correctly with `as`', () => {
        expect(
            render(
                <H2 alignment="left" mode="single-line" as="p">
                    This is a p
                </H2>
            ).asFragment()
        ).toMatchSnapshot()
    })
})
