import React from 'react'
import ReactDOM from 'react-dom'
import { act } from 'react-dom/test-utils'
import { of } from 'rxjs'
import { ActivationStatus } from './Activation'
import { ActivationClickTarget } from './ActivationClickTarget'

// describe('ActivationClickTarget', () => {
//     let container: HTMLDivElement
//     beforeEach(() => {
//         container = document.createElement('div')
//         document.body.appendChild(container)
//     })
//     afterEach(() => {
//         document.body.removeChild(container)
//     })

//     test('not yet activated', () => {
//         const activation = new ActivationStatus(
//             [
//                 {
//                     id: 'step1',
//                     title: 'title1',
//                     detail: 'detail1',
//                     action: () => undefined,
//                 },
//             ],
//             () => of({})
//         )
//         act(() => {
//             ReactDOM.render(
//                 <ActivationClickTarget activation={activation} activationKeys={['step1']}>
//                     <button>Button</button>
//                 </ActivationClickTarget>,
//                 container
//             )
//         })

//         expect(container.querySelector('.first-use-button')!.className).toBe('first-use-button ')
//         activation.update(null) // activation status loaded
//         expect(container.querySelector('.first-use-button')!.className).toBe('first-use-button ')
//         act(() => {
//             container.querySelector('button')!.dispatchEvent(new MouseEvent('click', { bubbles: true }))
//         })
//         expect(container.querySelector('.first-use-button')!.className).toBe('first-use-button animate')
//     })

//     test('already activated', () => {
//         const activation = new ActivationStatus(
//             [
//                 {
//                     id: 'step1',
//                     title: 'title1',
//                     detail: 'detail1',
//                     action: () => undefined,
//                 },
//             ],
//             () => of({ step1: true })
//         )
//         act(() => {
//             ReactDOM.render(
//                 <ActivationClickTarget activation={activation} activationKeys={['step1']}>
//                     <button>Button</button>
//                 </ActivationClickTarget>,
//                 container
//             )
//         })

//         expect(container.querySelector('.first-use-button')!.className).toBe('first-use-button ')
//         activation.update(null) // activation status loaded
//         expect(container.querySelector('.first-use-button')!.className).toBe('first-use-button ')
//         act(() => {
//             container.querySelector('button')!.dispatchEvent(new MouseEvent('click', { bubbles: true }))
//         })
//         expect(container.querySelector('.first-use-button')!.className).toBe('first-use-button ')
//     })
// })
