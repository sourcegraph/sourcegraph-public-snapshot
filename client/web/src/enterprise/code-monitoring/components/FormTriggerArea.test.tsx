import { screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { act } from 'react-dom/test-utils'
import sinon from 'sinon'

import { renderWithBrandedContext } from '@sourcegraph/shared/src/testing'

import { FormTriggerArea } from './FormTriggerArea'

describe('FormTriggerArea', () => {
    let clock: sinon.SinonFakeTimers

    beforeAll(() => {
        clock = sinon.useFakeTimers()
    })

    afterAll(() => {
        clock.restore()
    })

    const testCases = [
        { query: '', patternTypeChecked: true, typeChecked: false, repoChecked: false, validChecked: false },
        { query: 'test', patternTypeChecked: true, typeChecked: false, repoChecked: false, validChecked: true },
        {
            query: 'test patternType:literal',
            patternTypeChecked: true,
            typeChecked: false,
            repoChecked: false,
            validChecked: true,
        },
        {
            query: 'test patternType:regexp',
            patternTypeChecked: true,
            typeChecked: false,
            repoChecked: false,
            validChecked: true,
        },
        {
            query: 'test patternType:structural',
            patternTypeChecked: false,
            typeChecked: false,
            repoChecked: false,
            validChecked: true,
        },
        {
            query: 'test type:repo',
            patternTypeChecked: true,
            typeChecked: false,
            repoChecked: false,
            validChecked: true,
        },
        {
            query: 'test type:diff',
            patternTypeChecked: true,
            typeChecked: true,
            repoChecked: false,
            validChecked: true,
        },
        {
            query: 'test type:commit',
            patternTypeChecked: true,
            typeChecked: true,
            repoChecked: false,
            validChecked: true,
        },
        {
            query: 'test repo:test',
            patternTypeChecked: true,
            typeChecked: false,
            repoChecked: true,
            validChecked: true,
        },
        {
            query: 'test repo:test type:diff',
            patternTypeChecked: true,
            typeChecked: true,
            repoChecked: true,
            validChecked: true,
        },
    ]

    for (const testCase of testCases) {
        test(`Correct checkboxes checked for query '${testCase.query}'`, () => {
            renderWithBrandedContext(
                <FormTriggerArea
                    query={testCase.query}
                    triggerCompleted={false}
                    onQueryChange={sinon.spy()}
                    setTriggerCompleted={sinon.spy()}
                    startExpanded={false}
                    isLightTheme={true}
                    isSourcegraphDotCom={false}
                />
            )
            userEvent.click(screen.getByTestId('trigger-button'))
            act(() => {
                clock.tick(600)
            })

            const patternTypeCheckbox = screen.getByTestId('patterntype-checkbox')
            if (testCase.patternTypeChecked) {
                expect(patternTypeCheckbox).toBeChecked()
            } else {
                expect(patternTypeCheckbox).not.toBeChecked()
            }

            const typeCheckbox = screen.getByTestId('type-checkbox')
            if (testCase.typeChecked) {
                expect(typeCheckbox).toBeChecked()
            } else {
                expect(typeCheckbox).not.toBeChecked()
            }

            const repoCheckbox = screen.getByTestId('repo-checkbox')
            if (testCase.repoChecked) {
                expect(repoCheckbox).toBeChecked()
            } else {
                expect(repoCheckbox).not.toBeChecked()
            }

            const validCheckbox = screen.getByTestId('valid-checkbox')
            if (testCase.validChecked) {
                expect(validCheckbox).toBeChecked()
            } else {
                expect(validCheckbox).not.toBeChecked()
            }
        })
    }

    test('Append patternType:literal if no patternType is present', async () => {
        const onQueryChange = sinon.spy()
        renderWithBrandedContext(
            <FormTriggerArea
                query=""
                triggerCompleted={false}
                onQueryChange={onQueryChange}
                setTriggerCompleted={sinon.spy()}
                startExpanded={false}
                isLightTheme={true}
                isSourcegraphDotCom={false}
            />
        )
        userEvent.click(screen.getByTestId('trigger-button'))

        const triggerInput = screen.getByTestId('trigger-query-edit')
        userEvent.click(triggerInput)

        await waitFor(() => expect(triggerInput.querySelector('textarea[role="textbox"]')).toBeInTheDocument())

        const textbox = triggerInput.querySelector('textarea[role="textbox"]')!
        userEvent.type(textbox, 'test type:diff repo:test')

        act(() => {
            clock.tick(600)
        })
        userEvent.click(screen.getByTestId('submit-trigger'))

        sinon.assert.calledOnceWithExactly(onQueryChange, 'test type:diff repo:test patternType:literal')
    })

    test('Do not append patternType:literal if patternType is present', async () => {
        const onQueryChange = sinon.spy()
        renderWithBrandedContext(
            <FormTriggerArea
                query=""
                triggerCompleted={false}
                onQueryChange={onQueryChange}
                setTriggerCompleted={sinon.spy()}
                startExpanded={false}
                isLightTheme={true}
                isSourcegraphDotCom={false}
            />
        )
        userEvent.click(screen.getByTestId('trigger-button'))

        const triggerInput = screen.getByTestId('trigger-query-edit')
        userEvent.click(triggerInput)

        await waitFor(() => expect(triggerInput.querySelector('textarea[role="textbox"]')).toBeInTheDocument())

        const textbox = triggerInput.querySelector('textarea[role="textbox"]')!
        userEvent.type(textbox, 'test patternType:regexp type:diff repo:test')
        act(() => {
            clock.tick(600)
        })
        userEvent.click(screen.getByTestId('submit-trigger'))

        sinon.assert.calledOnceWithExactly(onQueryChange, 'test patternType:regexp type:diff repo:test')
    })
})
