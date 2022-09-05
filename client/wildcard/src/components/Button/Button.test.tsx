import { renderWithBrandedContext } from '@sourcegraph/shared/src/testing'

import { Button } from './Button'
import { BUTTON_VARIANTS, BUTTON_SIZES } from './constants'

describe('Button', () => {
    it('renders a simple button correctly', () => {
        const { asFragment } = renderWithBrandedContext(<Button>Hello world</Button>)
        expect(asFragment()).toMatchSnapshot()
    })

    it('supports rendering as different elements', () => {
        const { asFragment } = renderWithBrandedContext(<Button as="div">I am a div</Button>)
        expect(asFragment()).toMatchSnapshot()
    })

    it.each(BUTTON_VARIANTS)("Renders variant '%s' correctly", variant => {
        const { asFragment } = renderWithBrandedContext(<Button variant={variant}>Hello world</Button>)
        expect(asFragment()).toMatchSnapshot()
    })

    it.each(BUTTON_SIZES)("Renders size '%s' correctly", size => {
        const { asFragment } = renderWithBrandedContext(<Button size={size}>Hello world</Button>)
        expect(asFragment()).toMatchSnapshot()
    })
})
