import { Markdown } from './Markdown'
import { renderWithBrandedContext } from '../../testing'

describe('Markdown', () => {
    it('renders', () => {
        const component = renderWithBrandedContext(<Markdown dangerousInnerHTML="hello" />)
        expect(component.asFragment()).toMatchSnapshot()
    })
})
