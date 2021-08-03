import { render } from '@testing-library/react'
import React from 'react'

import { Button } from './Button'
import { BUTTON_VARIANTS, BUTTON_SIZES } from './constants'

describe('Button', () => {
    it('renders a simple button correctly', () => {
        const { container } = render(<Button>Hello world</Button>)
        expect(container.firstChild).toMatchInlineSnapshot(`
            <button
              class="btn"
              type="button"
            >
              Hello world
            </button>
        `)
    })

    it('supports rendering as different elements', () => {
        const { container } = render(<Button as="a">I am a link</Button>)
        expect(container.firstChild).toMatchInlineSnapshot(`
            <a
              class="btn"
            >
              I am a link
            </a>
        `)
    })

    it.each(BUTTON_VARIANTS)("Renders variant '%s' correctly", variant => {
        const { container } = render(<Button variant={variant}>Hello world</Button>)
        expect(container.firstChild).toMatchSnapshot()
    })

    it.each(BUTTON_SIZES)("Renders size '%s' correctly", size => {
        const { container } = render(<Button size={size}>Hello world</Button>)
        expect(container.firstChild).toMatchSnapshot()
    })
})
