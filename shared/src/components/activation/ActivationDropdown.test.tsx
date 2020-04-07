import * as H from 'history'
import React from 'react'
import renderer from 'react-test-renderer'
import { noop } from 'rxjs'
import { Activation } from './Activation'
import { ActivationDropdown } from './ActivationDropdown'
import ReactDOM from 'react-dom'

jest.mock('mdi-react/CheckboxBlankCircleIcon', () => 'CheckboxBlankCircleIcon')
jest.mock('mdi-react/CheckIcon', () => 'CheckIcon')

describe('ActivationDropdown', () => {
    const baseActivation: Activation = {
        steps: [
            {
                id: 'ConnectedCodeHost',
                title: 'title1',
                detail: 'detail1',
                onClick: noop,
            },
            {
                id: 'EnabledRepository',
                title: 'title2',
                detail: 'detail2',
            },
        ],
        refetch: noop,
        update: noop,
    }
    const history = H.createMemoryHistory({ keyLength: 0 })
    test('render loading', () => {
        const component = renderer.create(<ActivationDropdown history={history} activation={baseActivation} />, {
            createNodeMock: reactElement => {
                console.log(reactElement)
                return document.createElement(reactElement.type as string)
            },
        })
        expect(component.toJSON()).toMatchSnapshot()
    })
    // test('render 0/2 completed', () => {
    //     const component = renderer.create(
    //         <ActivationDropdown history={history} activation={{ ...baseActivation, completed: {} }} />,
    //         {
    //             createNodeMock: () => ({ ownerDocument: document.implementation.createHTMLDocument() }),
    //         }
    //     )
    //     expect(component.toJSON()).toMatchSnapshot()
    // })
    // test('render 1/2 completed', () => {
    //     const component = renderer.create(
    //         <ActivationDropdown
    //             history={history}
    //             activation={{ ...baseActivation, completed: { ConnectedCodeHost: true } }}
    //         />,
    //         {
    //             createNodeMock: () => ({ ownerDocument: document.implementation.createHTMLDocument() }),
    //         }
    //     )
    //     expect(component.toJSON()).toMatchSnapshot()
    // })
})
