import React from 'react'

import { fireEvent, screen } from '@testing-library/react'
import { createMemoryHistory, createLocation } from 'history'
import { NEVER } from 'rxjs'

import { renderWithBrandedContext } from '@sourcegraph/shared/src/testing'
import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'

import { mockAuthenticatedUser, mockCodeMonitorFields } from '../testing/util'

import { CodeMonitorForm, CodeMonitorFormProps } from './CodeMonitorForm'

const PROPS: CodeMonitorFormProps = {
    history: createMemoryHistory(),
    location: createLocation('/code-monitoring/new'),
    onSubmit: () => NEVER,
    submitButtonLabel: '',
    authenticatedUser: mockAuthenticatedUser,
}

describe('CodeMonitorForm', () => {
    test('Uses trigger query when present', () => {
        renderWithBrandedContext(
            <MockedTestProvider>
                <CodeMonitorForm {...PROPS} triggerQuery="foo" />
            </MockedTestProvider>
        )
        expect(screen.getByTestId('trigger-query-edit')).toHaveValue('foo')
    })

    test('Submit button disabled if no actions are present', () => {
        const { getByTestId } = renderWithBrandedContext(
            <MockedTestProvider>
                <CodeMonitorForm {...PROPS} codeMonitor={mockCodeMonitorFields} />
            </MockedTestProvider>
        )

        fireEvent.click(getByTestId('form-action-toggle-email'))
        fireEvent.click(getByTestId('delete-action-email'))

        expect(getByTestId('submit-monitor')).toBeDisabled()
    })

    test('Submit button enabled if one action is present', () => {
        const { getByTestId } = renderWithBrandedContext(
            <MockedTestProvider>
                <CodeMonitorForm {...PROPS} codeMonitor={{ ...mockCodeMonitorFields, actions: { nodes: [] } }} />
            </MockedTestProvider>
        )
        fireEvent.click(getByTestId('form-action-toggle-email'))
        fireEvent.click(getByTestId('submit-action-email'))

        expect(getByTestId('submit-monitor')).toBeEnabled()
    })
})
