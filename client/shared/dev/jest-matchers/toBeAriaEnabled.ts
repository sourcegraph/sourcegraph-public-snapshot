import { printReceived, matcherHint } from 'jest-matcher-utils'

export function toBeAriaEnabled(element: HTMLElement): { pass: boolean; message: () => string } {
    const isElement = element instanceof Element

    if (!isElement) {
        throw new Error('You should run this matcher over a valid html element')
    }

    const pass = element.getAttribute('aria-disabled') !== 'true' && element.getAttribute('disabled') === null

    const passMessage = `${matcherHint('.not.toBeAriaDisabled', 'received', '')}
        Expected element should have an aria disabled state but received:
           aria-disabled: ${printReceived(element.ariaDisabled)}
          disabled: ${printReceived(element.getAttribute('disabled'))}
    `
    const failMessage = `${matcherHint('.toBeArray', 'received', '')}
        Expected element should not have an aria disabled state but received:
          aria-disabled: ${printReceived(element.ariaDisabled)}
          disabled: ${printReceived(element.getAttribute('disabled'))}
    `

    return {
        pass,
        message: () => (pass ? passMessage : failMessage),
    }
}
