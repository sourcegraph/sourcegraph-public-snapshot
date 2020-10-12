import { noop } from 'lodash'
import { Observable, Subject } from 'rxjs'
import * as sinon from 'sinon'
import { createValidationPipeline, InputValidationState } from './useInputValidation'

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
        const internalStrings = {
            value: '',
            validationMessage: '',
        }

        return {
            get value() {
                return internalStrings.value
            },
            get validationMessage() {
                return internalStrings.validationMessage
            },
            checkValidity() {
                // Check if custom validity was set
                if (internalStrings.validationMessage.length > 0) {
                    return false
                }
                // Built-in rules
                if (!internalStrings.value.includes('@')) {
                    internalStrings.validationMessage = "Email must include '@'"
                    return false
                }
                // Clear built-in messages
                internalStrings.validationMessage = ''
                return true
            },
            setCustomValidity(error) {
                internalStrings.validationMessage = error
            },
            changeValue(newValue) {
                internalStrings.value = newValue
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
    it.skip('works without initial value', () => {
        const inputElement = createEmailInputElement()
        const inputReference = { current: inputElement }

        function isDotCo(email: string): string | undefined {
            if (email.endsWith('.co')) {
                return undefined
            }

            return "Email must end with '.co'"
        }

        const validationStates: Subject<InputValidationState> = new Subject()

        const validationPipeline = createValidationPipeline(
            {
                synchronousValidators: [isDotCo],
            },
            inputReference,
            // We want to test the value that this callback is called with,
            // not the emissions of the returned observable. Therefore, we will
            // redirect these values to another observable whose emissions we will assert.
            validationStates.next.bind(validationStates)
        )

        // Creating this type instead of a generic util because TS doesn't support higher-kinded types
        type ObservableEmission<T> = T extends Observable<infer V> ? V : never
        const changeEvents = new Subject<ObservableEmission<Parameters<typeof validationPipeline>[0]>>()

        // This is not the observable that we care about. Here, we simply set up the pipeline
        // change event -> validation pipeline -> validation states
        const validationResults = validationPipeline(changeEvents)

        validationResults.subscribe(value => console.log('obsval', value))
        validationStates.subscribe(value => console.log('valstates', value))

        // Simulate user input: change input value, then dispatch change event
        inputElement.changeValue('so')
        changeEvents.next({
            preventDefault: noop,
            target: inputReference.current,
        })

        inputElement.changeValue('so@so.co')
        changeEvents.next({
            preventDefault: noop,
            target: inputReference.current,
        })

        console.log('value in test', inputReference.current.value)

        // return new Promise(resolve => setTimeout(resolve, 1200))
        clock.tick(1200)
    })
})
