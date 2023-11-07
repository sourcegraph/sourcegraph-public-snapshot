import { afterEach, beforeEach, describe, expect, test } from '@jest/globals'
import { fireEvent, getByRole, screen } from '@testing-library/react'
import { NEVER } from 'rxjs'

import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'
import { assertAriaDisabled } from '@sourcegraph/testing'
import { renderWithBrandedContext } from '@sourcegraph/wildcard/src/testing'

import { mockAuthenticatedUser, mockCodeMonitorFields } from '../testing/util'

import { CodeMonitorForm, type CodeMonitorFormProps } from './CodeMonitorForm'

const PROPS: CodeMonitorFormProps = {
    onSubmit: () => NEVER,
    submitButtonLabel: '',
    authenticatedUser: mockAuthenticatedUser,
    isSourcegraphDotCom: false,
}

describe('CodeMonitorForm', () => {
    const origContext = window.context
    beforeEach(() => {
        window.context = {
            emailEnabled: true,
        } as any
    })
    afterEach(() => {
        window.context = origContext
    })

    test('Uses trigger query when present', () => {
        renderWithBrandedContext(
            <MockedTestProvider>
                <CodeMonitorForm {...PROPS} triggerQuery="foo" />
            </MockedTestProvider>
        )

        const triggerEdit = screen.getByTestId('trigger-query-edit')
        expect(getByRole(triggerEdit, 'textbox')).toHaveValue('foo')
    })

    test('Submit button disabled if no actions are present', () => {
        const { getByTestId } = renderWithBrandedContext(
            <MockedTestProvider>
                <CodeMonitorForm {...PROPS} codeMonitor={mockCodeMonitorFields} />
            </MockedTestProvider>
        )

        fireEvent.click(getByTestId('form-action-toggle-email'))
        fireEvent.click(getByTestId('delete-action-email'))

        assertAriaDisabled(getByTestId('submit-monitor'))
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
