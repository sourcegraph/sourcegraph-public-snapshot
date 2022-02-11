import { fireEvent, screen } from '@testing-library/react'
import { createMemoryHistory, createLocation } from 'history'
import React from 'react'
import { NEVER } from 'rxjs'

import { renderWithBrandedContext } from '@sourcegraph/shared/src/testing'

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
        renderWithBrandedContext(<CodeMonitorForm {...PROPS} triggerQuery="foo" />)
        expect(screen.getByTestId('trigger-query-edit')).toHaveValue('foo')
    })

    test('Submit button disabled if no actions are present', () => {
        const { getByTestId } = renderWithBrandedContext(
            <CodeMonitorForm {...PROPS} codeMonitor={mockCodeMonitorFields} />
        )

        fireEvent.click(getByTestId('form-action-toggle-email'))
        fireEvent.click(getByTestId('delete-action-email'))

        expect(getByTestId('submit-monitor')).toBeDisabled()
    })

    test('Submit button enabled if one action is present', () => {
        const { getByTestId } = renderWithBrandedContext(
            <CodeMonitorForm {...PROPS} codeMonitor={{ ...mockCodeMonitorFields, actions: { nodes: [] } }} />
        )
        fireEvent.click(getByTestId('form-action-toggle-email'))
        fireEvent.click(getByTestId('submit-action-email'))

        expect(getByTestId('submit-monitor')).toBeEnabled()
    })
})
