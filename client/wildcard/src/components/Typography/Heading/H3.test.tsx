import { render } from '@testing-library/react'

import { Typography } from '../..'

describe('H3', () => {
    it('should render correctly', () => {
        expect(
            render(
                <Typography.H3 alignment="left" mode="single-line">
                    This is H3
                </Typography.H3>
            ).asFragment()
        ).toMatchSnapshot()
    })

    it('should render correctly with `as`', () => {
        expect(
            render(
                <Typography.H3 alignment="left" mode="single-line" as="p">
                    This is a p
                </Typography.H3>
            ).asFragment()
        ).toMatchSnapshot()
    })
})
