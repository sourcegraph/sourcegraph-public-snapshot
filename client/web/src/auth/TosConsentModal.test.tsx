import { MockedProvider } from '@apollo/client/testing'
import { fireEvent, render } from '@testing-library/react'
import * as sinon from 'sinon'

import { MockedTestProvider, waitForNextApolloResponse } from '@sourcegraph/shared/src/testing/apollo'

import { SET_TOS_ACCEPTED_MUTATION, TosConsentModal } from './TosConsentModal'

describe('TosConsentModal', () => {
    it('should call afterTosAccepted if request was successful', async () => {
        const mocks = [
            {
                request: {
                    query: SET_TOS_ACCEPTED_MUTATION,
                },
                result: { data: { alwaysNil: null } },
            },
        ]

        const afterTosAccepted = sinon.spy()
        const component = render(
            <MockedProvider mocks={mocks}>
                <TosConsentModal afterTosAccepted={afterTosAccepted} />
            </MockedProvider>
        )

        const checkbox = component.getByLabelText(/I agree/)
        fireEvent.click(checkbox)

        const button = component.getByText(/Agree and continue/)
        fireEvent.click(button)

        await waitForNextApolloResponse()

        expect(afterTosAccepted.calledOnce).toBeTruthy()
    })

    it('should not call afterTosAccepted if there was an error', async () => {
        const mocks = [
            {
                request: {
                    query: SET_TOS_ACCEPTED_MUTATION,
                },
                error: new Error('An error occurred'),
            },
        ]

        const afterTosAccepted = sinon.spy()
        const component = render(
            <MockedTestProvider mocks={mocks}>
                <TosConsentModal afterTosAccepted={afterTosAccepted} />
            </MockedTestProvider>
        )

        const checkbox = component.getByLabelText(/I agree/)
        fireEvent.click(checkbox)

        const button = component.getByText(/Agree and continue/)
        fireEvent.click(button)

        await waitForNextApolloResponse()

        expect(afterTosAccepted.notCalled).toBeTruthy()
    })
})
