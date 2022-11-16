import { expect } from '@jest/globals'
import { printReceived, matcherHint } from 'jest-matcher-utils'

// https://github.com/testing-library/jest-dom/issues/144
export function toBeAriaDisabled(element: HTMLElement): { pass: boolean; message: () => string } {
    const isElement = element instanceof Element

    if (!isElement) {
        throw new Error('You should run this matcher over a valid html element')
    }

    const pass = element.getAttribute('aria-disabled') === 'true'

    const passMessage = `${matcherHint('.not.toBeAriaDisabled', 'received', '')}
        Expected element should not have an aria disabled state but received:
          ${printReceived(element.ariaDisabled)}
    `
    const failMessage = `${matcherHint('.toBeArray', 'received', '')}
        Expected element should have an aria disabled state but received:
          ${printReceived(element.ariaDisabled)}
    `

    return {
        pass,
        message: () => (pass ? passMessage : failMessage),
    }
}

expect.extend({ toBeAriaDisabled })

declare global {
    // eslint-disable-next-line @typescript-eslint/no-namespace
    namespace jest {
        interface Matchers<R, T> {
            toBeAriaDisabled(): R
        }
    }
}

