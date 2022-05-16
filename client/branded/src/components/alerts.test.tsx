import { render } from '@testing-library/react'

import { ErrorAlert } from './alerts'

jest.mock('mdi-react/AlertCircleIcon', () => 'AlertCircleIcon')

describe('ErrorAlert', () => {
    it('should render an Error object as an alert', () => {
        expect(render(<ErrorAlert error={new Error('an error happened')} />).asFragment()).toMatchSnapshot()
    })

    it('should add a prefix if given', () => {
        expect(
            render(<ErrorAlert error={new Error('an error happened')} prefix="An error happened" />).asFragment()
        ).toMatchSnapshot()
    })

    it('should omit the icon if icon={false}', () => {
        expect(
            render(<ErrorAlert error={new Error('an error happened')} icon={false} />).asFragment()
        ).toMatchSnapshot()
    })

    it('should render a Go multierror nicely', () => {
        expect(
            render(
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
