import { describe, expect, it } from 'vitest'

import { renderWithBrandedContext } from '../../testing'

import { Markdown } from './Markdown'

describe('Markdown', () => {
    it('renders', () => {
        const component = renderWithBrandedContext(<Markdown dangerousInnerHTML="hello" />)
        expect(component.asFragment()).toMatchSnapshot()
    })
})
