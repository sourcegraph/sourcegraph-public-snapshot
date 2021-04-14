import { mount } from 'enzyme'
import React from 'react'
import { act } from 'react-dom/test-utils'
import sinon from 'sinon'

import { FormTriggerArea } from './FormTriggerArea'

describe('FormTriggerArea', () => {
    let clock: sinon.SinonFakeTimers

    beforeAll(() => {
        clock = sinon.useFakeTimers()
    })

    afterAll(() => {
        clock.restore()
    })

    test('Correct checkboxes shown when query does not fulfill requirements', () => {
        let component = mount(
            <FormTriggerArea
                query="test repo:test"
                triggerCompleted={false}
                onQueryChange={sinon.spy()}
                setTriggerCompleted={sinon.spy()}
                startExpanded={false}
            />
        )
        act(() => {
            const triggerButton = component.find('.test-trigger-button')
            triggerButton.simulate('click')
            clock.tick(600)
        })
        component = component.update()

        expect(component).toMatchSnapshot()
    })

    const testCases = [
        { query: '', typeChecked: false, repoChecked: false, validChecked: false },
        { query: 'test', typeChecked: false, repoChecked: false, validChecked: true },
        { query: 'test type:repo', typeChecked: false, repoChecked: false, validChecked: true },
        { query: 'test type:diff', typeChecked: true, repoChecked: false, validChecked: true },
        { query: 'test type:commit', typeChecked: true, repoChecked: false, validChecked: true },
        { query: 'test repo:test', typeChecked: false, repoChecked: true, validChecked: true },
        { query: 'test repo:test type:diff', typeChecked: true, repoChecked: true, validChecked: true },
    ]

    for (const testCase of testCases) {
        test(`Correct checkboxes checked for query '${testCase.query}'`, () => {
            let component = mount(
                <FormTriggerArea
                    query={testCase.query}
                    triggerCompleted={false}
                    onQueryChange={sinon.spy()}
                    setTriggerCompleted={sinon.spy()}
                    startExpanded={false}
                />
            )
            act(() => {
                const triggerButton = component.find('.test-trigger-button')
                triggerButton.simulate('click')
                clock.tick(600)
            })
            component = component.update()

            const typeCheckbox = component.find('.test-type-checkbox input[type="checkbox"]')
            expect(typeCheckbox.get(0).props?.checked).toBe(testCase.typeChecked)

            const repoCheckbox = component.find('.test-repo-checkbox input[type="checkbox"]')
            expect(repoCheckbox.get(0).props?.checked).toBe(testCase.repoChecked)

            const validCheckbox = component.find('.test-valid-checkbox input[type="checkbox"]')
            expect(validCheckbox.get(0).props?.checked).toBe(testCase.validChecked)
        })
    }

    test('Append patternType:literal if no patternType is present', () => {
        const onQueryChange = sinon.spy()
        let component = mount(
            <FormTriggerArea
                query=""
                triggerCompleted={false}
                onQueryChange={onQueryChange}
                setTriggerCompleted={sinon.spy()}
                startExpanded={false}
            />
        )
        const triggerButton = component.find('.test-trigger-button')
        triggerButton.simulate('click')

        act(() => {
            const triggerInput = component.find('.test-trigger-input')
            triggerInput.simulate('change', { target: { value: 'test type:diff repo:test' } })
            clock.tick(600)
        })
        component = component.update()
        const submitButton = component.find('.test-submit-trigger')
        submitButton.simulate('click')

        sinon.assert.calledOnceWithExactly(onQueryChange, 'test type:diff repo:test patternType:literal')
    })

    test('Do not append patternType:literal if no patternType is present', () => {
        const onQueryChange = sinon.spy()
        let component = mount(
            <FormTriggerArea
                query=""
                triggerCompleted={false}
                onQueryChange={onQueryChange}
                setTriggerCompleted={sinon.spy()}
                startExpanded={false}
            />
        )
        const triggerButton = component.find('.test-trigger-button')
        triggerButton.simulate('click')

        act(() => {
            const triggerInput = component.find('.test-trigger-input')
            triggerInput.simulate('change', { target: { value: 'test patternType:regexp type:diff repo:test' } })
            clock.tick(600)
        })
        component = component.update()
        const submitButton = component.find('.test-submit-trigger')
        submitButton.simulate('click')

        sinon.assert.calledOnceWithExactly(onQueryChange, 'test patternType:regexp type:diff repo:test')
    })
})
