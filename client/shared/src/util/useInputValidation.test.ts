import { noop } from 'lodash'
import { Observable, Subject } from 'rxjs'
import * as sinon from 'sinon'
import { createValidationPipeline } from './useInputValidation'

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

    /**
     * Marble tests
     * - Works without initial value
     * - Works with initial value
     * - Built-in sync reason
     * - Custom sync reason
     * - Async reason
     * - All validators passed
     */
    it('works without initial value', () => {
        const { changeValue, ...inputElement } = createEmailInputElement()
        const inputReference = { current: inputElement }

        function isDotCo(email: string): string | undefined {
            if (email.endsWith('.co')) {
                return undefined
            }

            return "Email must end with '.co'"
        }

        const validationPipeline = createValidationPipeline(
            {
                synchronousValidators: [isDotCo],
            },
            inputReference,
            () => {}
        )

        // Creating this type instead of a generic util because TS doesn't support higher-kinded types
        type ObservableEmission<T> = T extends Observable<infer V> ? V : never
        const changeEvents = new Subject<ObservableEmission<Parameters<typeof validationPipeline>[0]>>()

        const validationResults = validationPipeline(changeEvents)

        changeEvents.next({
            preventDefault: noop,
            target: inputReference.current,
        })
    })
})
