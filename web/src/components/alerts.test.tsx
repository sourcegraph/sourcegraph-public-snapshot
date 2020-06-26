import { ErrorAlert } from './alerts'
import React from 'react'
import { createMemoryHistory } from 'history'
import { mount } from 'enzyme'

jest.mock('mdi-react/AlertCircleIcon', () => 'AlertCircleIcon')

describe('ErrorAlert', () => {
    it('should render an Error object as an alert', () => {
        expect(
            mount(<ErrorAlert error={new Error('an error happened')} history={createMemoryHistory()} />).children()
        ).toMatchSnapshot()
    })

    it('should add a prefix if given', () => {
        expect(
            mount(
                <ErrorAlert
                    error={new Error('an error happened')}
                    prefix="An error happened"
                    history={createMemoryHistory()}
                />
            ).children()
        ).toMatchSnapshot()
    })

    it('should omit the icon if icon={false}', () => {
        expect(
            mount(
                <ErrorAlert error={new Error('an error happened')} icon={false} history={createMemoryHistory()} />
            ).children()
        ).toMatchSnapshot()
    })

    it('should render a Go multierror nicely', () => {
        expect(
            mount(
                <ErrorAlert
                    error={
                        new Error(
                            '- Additional property asdasd is not allowed\n- projectQuery.0: String length must be greater than or equal to 1\n'
                        )
                    }
                    history={createMemoryHistory()}
                />
            ).children()
        ).toMatchSnapshot()
    })
})
