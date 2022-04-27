import { render } from '@testing-library/react'

import { Markdown } from './Markdown'

describe('Markdown', () => {
    it('renders', () => {
        const component = render(<Markdown dangerousInnerHTML="hello" />)
        expect(component.asFragment()).toMatchSnapshot()
    })
})
