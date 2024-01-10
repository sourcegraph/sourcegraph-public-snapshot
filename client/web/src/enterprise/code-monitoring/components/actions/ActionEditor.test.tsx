import { render } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import sinon from 'sinon'
import { describe, expect, test } from 'vitest'

import { assertAriaDisabled } from '@sourcegraph/testing'

import { ActionEditor, type ActionEditorProps } from './ActionEditor'

describe('ActionEditor', () => {
    const props: ActionEditorProps = {
        title: 'Send email notifications',
        subtitle: 'Send notifications to specified recipients.',
        disabled: false,
        completed: false,
        completedSubtitle: '',
        idName: 'email',
        actionEnabled: true,
        toggleActionEnabled: sinon.fake(),
        includeResults: true,
        toggleIncludeResults: sinon.fake(),
        canSubmit: true,
        onSubmit: sinon.fake(),
        onCancel: sinon.fake(),
        canDelete: true,
        onDelete: sinon.fake(),
        testButtonText: 'Send test email',
        onTest: sinon.fake(),
        testAgainButtonText: 'Send again',
        testState: undefined,
    }

    test('expand and collapse with cancel button', () => {
        const cancelSpy = sinon.spy()

        const { getByTestId } = render(<ActionEditor {...props} onCancel={cancelSpy} />)
        userEvent.click(getByTestId('form-action-toggle-email'))

        expect(getByTestId('cancel-action-email')).toBeVisible()

        userEvent.click(getByTestId('cancel-action-email'))
        expect(getByTestId('form-action-toggle-email')).toBeVisible()
        sinon.assert.calledOnce(cancelSpy)
    })

    test('expand and collapse with cancel button, completed', () => {
        const cancelSpy = sinon.spy()

        const { getByTestId } = render(<ActionEditor {...props} onCancel={cancelSpy} completed={true} />)
        userEvent.click(getByTestId('form-action-toggle-email'))

        expect(getByTestId('cancel-action-email')).toBeVisible()

        userEvent.click(getByTestId('cancel-action-email'))
        expect(getByTestId('form-action-toggle-email')).toBeVisible()
        sinon.assert.calledOnce(cancelSpy)
    })

    test('expand and collapse with submit button', () => {
        const submitSpy = sinon.spy()

        const { getByTestId } = render(<ActionEditor {...props} onSubmit={submitSpy} />)
        userEvent.click(getByTestId('form-action-toggle-email'))

        expect(getByTestId('submit-action-email')).toBeVisible()

        userEvent.click(getByTestId('submit-action-email'))
        expect(getByTestId('form-action-toggle-email')).toBeVisible()
        sinon.assert.calledOnce(submitSpy)
    })

    test('expand and collapse with delete button', () => {
        const deleteSpy = sinon.spy()

        const { getByTestId } = render(<ActionEditor {...props} completed={true} onDelete={deleteSpy} />)
        userEvent.click(getByTestId('form-action-toggle-email'))

        expect(getByTestId('delete-action-email')).toBeVisible()

        userEvent.click(getByTestId('delete-action-email'))
        expect(getByTestId('form-action-toggle-email')).toBeVisible()
        sinon.assert.calledOnce(deleteSpy)
    })

    test('submit and delete disabled', () => {
        const { getByTestId, queryByTestId } = render(<ActionEditor {...props} canSubmit={false} canDelete={false} />)
        userEvent.click(getByTestId('form-action-toggle-email'))

        expect(queryByTestId('delete-action-email')).not.toBeInTheDocument()
        assertAriaDisabled(getByTestId('submit-action-email'))
    })

    test('toggle disable when collapsed', () => {
        const toggleActionEnabledSpy = sinon.spy()
        const { getByTestId } = render(
            <ActionEditor
                {...props}
                completed={true}
                actionEnabled={true}
                toggleActionEnabled={toggleActionEnabledSpy}
            />
        )
        expect(getByTestId('enable-action-toggle-collapsed-email')).toBeChecked()

        userEvent.click(getByTestId('enable-action-toggle-collapsed-email'))
        sinon.assert.calledWithExactly(toggleActionEnabledSpy, false, true)
    })

    test('toggle disable when expanded', () => {
        const toggleActionEnabledSpy = sinon.spy()
        const { getByTestId } = render(
            <ActionEditor
                {...props}
                completed={false}
                actionEnabled={false}
                toggleActionEnabled={toggleActionEnabledSpy}
            />
        )
        userEvent.click(getByTestId('form-action-toggle-email'))
        expect(getByTestId('enable-action-toggle-expanded-email')).not.toBeChecked()

        userEvent.click(getByTestId('enable-action-toggle-expanded-email'))
        sinon.assert.calledWithExactly(toggleActionEnabledSpy, true, false)
    })
})
