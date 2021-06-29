import React from 'react'
import renderer from 'react-test-renderer'

import { ActivationChecklist } from './ActivationChecklist'

jest.mock('mdi-react/CheckboxBlankCircleIcon', () => 'CheckboxBlankCircleIcon')
jest.mock('mdi-react/CheckIcon', () => 'CheckIcon')

describe('ActivationChecklist', () => {
    test('render loading', () => {
        const component = renderer.create(<ActivationChecklist steps={[]} />)
        expect(component.toJSON()).toMatchSnapshot()
    })
    test('render 0/1 complete', () => {
        {
            const component = renderer.create(
                <ActivationChecklist
                    steps={[
                        {
                            id: 'ConnectedCodeHost',
                            title: 'title1',
                            detail: 'detail1',
                        },
                    ]}
                    completed={{}}
                />
            )
            expect(component.toJSON()).toMatchSnapshot()
        }
        {
            const component = renderer.create(
                <ActivationChecklist
                    steps={[
                        {
                            id: 'ConnectedCodeHost',
                            title: 'title1',
                            detail: 'detail1',
                        },
                    ]}
                    completed={{ EnabledRepository: true }} // another item
                />
            )
            expect(component.toJSON()).toMatchSnapshot()
        }
    })
    test('render 1/1 complete', () => {
        const component = renderer.create(
            <ActivationChecklist
                steps={[
                    {
                        id: 'ConnectedCodeHost',
                        title: 'title1',
                        detail: 'detail1',
                    },
                ]}
                completed={{ ConnectedCodeHost: true }} // same item as in steps
            />
        )
        expect(component.toJSON()).toMatchSnapshot()
    })
})
