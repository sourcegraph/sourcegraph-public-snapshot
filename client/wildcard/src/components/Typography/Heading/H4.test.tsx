import { render } from '@testing-library/react'
import { describe, expect, it } from 'vitest'

import { H4 } from './H4'

describe('H4', () => {
    it('should render correctly', () => {
        expect(
            render(
                <H4 alignment="left" mode="single-line">
                    This is H4
                </H4>
            ).asFragment()
        ).toMatchSnapshot()
    })

    it('should render correctly with `as`', () => {
        expect(
            render(
                <H4 alignment="left" mode="single-line" as="p">
                    This is a p
                </H4>
            ).asFragment()
        ).toMatchSnapshot()
    })
})
