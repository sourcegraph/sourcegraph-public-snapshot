import { act, renderHook } from '@testing-library/react-hooks'
import * as sinon from 'sinon'
import { useInputValidation } from './useInputValidation'

describe('useInputValidation()', () => {
    let clock: sinon.SinonFakeTimers
    beforeAll(() => {
        clock = sinon.useFakeTimers()
    })

    afterAll(() => {
        clock.restore()
    })

    /**
     * Creates a mock input element for emails. Only checks for '@'. Acts as
     * if `novalidate` is true.
     *
     * Implements the `HTMLInputElement` properties needed by `useInputValidation`,
     * along with a `changeValue` method to simulate input and change the internal value
     */
    function createEmailInputElement(): Pick<
        HTMLInputElement,
        'checkValidity' | 'validationMessage' | 'setCustomValidity' | 'value'
    > & { changeValue: (newValue: string) => void } {
        let value = ''
        let validationMessage = ''

        return {
            validationMessage,
            checkValidity: () => {
                // Check if custom validity was set
                if (validationMessage.length > 0) {
                    return false
                }
                // Built-in rules
                if (!value.includes('@')) {
                    validationMessage = "Email must include '@'"
                    return false
                }
                // Clear built-in messages
                validationMessage = ''
                return true
            },
            setCustomValidity: error => {
                validationMessage = error
            },
            changeValue: newValue => {
                console.log('changin value', value)
                value = newValue
                console.log('changed value', value)
            },
            get value() {
                return value
            },
        }
    }

    it.skip('works with custom sync validators', () => {
        function isDotCo(email: string): string | undefined {
            if (email.endsWith('.co')) {
                return undefined
            }

            return "Email must end with '.co'"
        }

        const { result } = renderHook(() =>
            useInputValidation({
                synchronousValidators: [isDotCo],
            })
        )

        const { changeValue, ...inputElement } = createEmailInputElement()
        result.current[2].current = inputElement as HTMLInputElement

        function simulateUserInput(value: string): void {
            // Call `changeValue` before `onChange` so that the internal value is accurate
            changeValue(value)
            // eslint-disable-next-line @typescript-eslint/no-floating-promises
            act(() => {
                // eslint-disable-next-line @typescript-eslint/consistent-type-assertions
                result.current[1]({
                    preventDefault: () => {},
                    target: inputElement as EventTarget & HTMLInputElement,
                } as React.ChangeEvent<HTMLInputElement>)
            })
        }

        // Valid email
        simulateUserInput('sourcegraph@sg.co')

        clock.tick(1200)
        console.log('state?', result.current[0])

        expect(true).toBe(true)
    })

    it.skip('works with async validators', () => {})
})
