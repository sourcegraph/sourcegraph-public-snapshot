import { render } from '@testing-library/react'
import { describe, expect, it } from 'vitest'

import { H1 } from './H1'

describe('H1', () => {
    it('should render correctly', () => {
        expect(
            render(
                <H1 alignment="left" mode="single-line">
                    This is H1
                </H1>
            ).asFragment()
        ).toMatchSnapshot()
    })

    it('should render correctly with `as`', () => {
        expect(
            render(
                <H1 alignment="left" mode="single-line" as="p">
                    This is a p
                </H1>
            ).asFragment()
        ).toMatchSnapshot()
    })
})
