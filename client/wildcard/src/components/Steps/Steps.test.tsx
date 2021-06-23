import { render, RenderResult, cleanup /* fireEvent */ } from '@testing-library/react'
import React from 'react'
import sinon from 'sinon'

import { Steps, Step, StepsProps } from './Steps'

describe('Steps', () => {
    // let queries: RenderResult
    const renderWithProps = (props: StepsProps): RenderResult =>
        render(
            <Steps {...props}>
                <Step title="Connect with code hosts" />
                <Step title="Add Repositories" />
                <Step title="Start Searching" />
            </Steps>
        )
    const onChangeMock = sinon.spy()

    beforeEach(() => {
        onChangeMock.resetHistory()
    })

    afterEach(cleanup)

    describe('Invalid configuration', () => {
        it('will error when initial step is less than 1', () => {
            expect(() => {
                renderWithProps({ current: 1, initial: 0, onChange: onChangeMock })
            }).toThrowErrorMatchingSnapshot()
        })
    })
})
