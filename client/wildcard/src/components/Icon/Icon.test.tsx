import { mdiCheck } from '@mdi/js'
import { render } from '@testing-library/react'
import CheckIcon from 'mdi-react/CheckIcon'

import { Icon } from './Icon'

import { AccessibleSvg } from '.'

const CustomIcon: AccessibleSvg = ({ className, size, ...props }) => (
    <svg height={size} width={size} className={className} viewBox="0 0 24 24" {...props}>
        {'title' in props ? <title>{props.title}</title> : null}
        <path d="M21,7L9,19L3.5,13.5L4.91,12.09L9,16.17L19.59,5.59L21,7Z" />
    </svg>
)

describe('Icon', () => {
    describe('mdi-react icons (legacy)', () => {
        it('renders a simple icon correctly', () => {
            const { asFragment } = render(<Icon as={CheckIcon} />)
            expect(asFragment().firstChild).toMatchInlineSnapshot(`
                <svg
                  class="mdi-icon iconInline"
                  fill="currentColor"
                  height="24"
                  viewBox="0 0 24 24"
                  width="24"
                >
                  <path
                    d="M21,7L9,19L3.5,13.5L4.91,12.09L9,16.17L19.59,5.59L21,7Z"
                  />
                </svg>
            `)
        })
    })

    describe('@mdi/react icons (current)', () => {
        it('renders a simple icon correctly', () => {
            const { asFragment } = render(<Icon svgPath={mdiCheck} aria-hidden="true" />)
            expect(asFragment().firstChild).toMatchInlineSnapshot(`
                <svg
                  aria-hidden="true"
                  class="iconInline"
                  role="presentation"
                  viewBox="0 0 24 24"
                >
                  <path
                    d="M21,7L9,19L3.5,13.5L4.91,12.09L9,16.17L19.59,5.59L21,7Z"
                    style="fill: currentColor;"
                  />
                </svg>
            `)
        })

        it('renders a simple icon that can be understood by a screen reader correctly', () => {
            const { asFragment } = render(<Icon svgPath={mdiCheck} title="Checkmark" />)
            expect(asFragment().firstChild).toMatchInlineSnapshot(`
                <svg
                  aria-labelledby="icon_labelledby_2"
                  class="iconInline"
                  viewBox="0 0 24 24"
                >
                  <title
                    id="icon_labelledby_2"
                  >
                    Checkmark
                  </title>
                  <path
                    d="M21,7L9,19L3.5,13.5L4.91,12.09L9,16.17L19.59,5.59L21,7Z"
                    style="fill: currentColor;"
                  />
                </svg>
            `)
        })
    })

    describe('custom SVGs that implement the AccessibleSvg type', () => {
        it('renders a simple icon correctly', () => {
            const { asFragment } = render(<Icon as={CustomIcon} aria-hidden="true" />)
            expect(asFragment().firstChild).toMatchInlineSnapshot(`
                <svg
                  aria-hidden="true"
                  class="iconInline"
                  viewBox="0 0 24 24"
                >
                  <path
                    d="M21,7L9,19L3.5,13.5L4.91,12.09L9,16.17L19.59,5.59L21,7Z"
                  />
                </svg>
            `)
        })

        it('renders a simple icon that can be understood by a screen reader correctly', () => {
            const { asFragment } = render(<Icon as={CustomIcon} title="Sourcegraph icon" />)
            expect(asFragment().firstChild).toMatchInlineSnapshot(`
                <svg
                  class="iconInline"
                  title="Sourcegraph icon"
                  viewBox="0 0 24 24"
                >
                  <title>
                    Sourcegraph icon
                  </title>
                  <path
                    d="M21,7L9,19L3.5,13.5L4.91,12.09L9,16.17L19.59,5.59L21,7Z"
                  />
                </svg>
            `)
        })
    })
})
