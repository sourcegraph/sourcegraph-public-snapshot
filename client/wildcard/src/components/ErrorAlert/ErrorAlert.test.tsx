import { describe, expect, it } from 'vitest'

import { renderWithBrandedContext } from '../../testing'

import { ErrorAlert } from './ErrorAlert'

describe('ErrorAlert', () => {
    it('should render an Error object as an alert', () => {
        expect(
            renderWithBrandedContext(<ErrorAlert error={new Error('an error happened')} />).asFragment()
        ).toMatchSnapshot()
    })

    it('should add a prefix if given', () => {
        expect(
            renderWithBrandedContext(
                <ErrorAlert error={new Error('an error happened')} prefix="An error happened" />
            ).asFragment()
        ).toMatchSnapshot()
    })

    it('should render a Go multierror nicely', () => {
        expect(
            renderWithBrandedContext(
                <ErrorAlert
                    error={
                        new Error(
                            '- Additional property asdasd is not allowed\n- projectQuery.0: String length must be greater than or equal to 1\n'
                        )
                    }
                />
            ).asFragment()
        ).toMatchSnapshot()
    })
})
