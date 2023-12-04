import { render } from '@testing-library/react'
import { describe, expect, it } from 'vitest'

import { BaseControlInput, BASE_CONTROL_TYPES } from './BaseControlInput'

describe('BaseControlInput', () => {
    describe.each(BASE_CONTROL_TYPES)('%s input', type => {
        it('renders correctly with a visible label', () => {
            const { container } = render(<BaseControlInput id="test" type={type} label="Hello world" />)
            expect(container.firstChild).toMatchSnapshot()
        })

        it('renders correctly with a hidden label', () => {
            const { container } = render(<BaseControlInput id="test" type={type} aria-label="Hello world" />)
            expect(container.firstChild).toMatchSnapshot()
        })

        it('renders correctly with an externally provided label', () => {
            const { container } = render(
                <div>
                    <span id="test-label">Hello world</span>
                    <BaseControlInput id="test" type={type} aria-labelledby="test-label" />
                </div>
            )
            expect(container.firstChild).toMatchSnapshot()
        })

        it('renders with message correctly', () => {
            const { container } = render(
                <BaseControlInput id="test" type={type} label="Hello world" message="Additional message" />
            )
            expect(container.firstChild).toMatchSnapshot()
        })

        it('renders with correct styles when invalid', () => {
            const { container } = render(
                <BaseControlInput
                    id="test"
                    type={type}
                    label="Hello world"
                    message="Additional message"
                    isValid={false}
                />
            )
            expect(container.firstChild).toMatchSnapshot()
        })

        it('renders with correct styles when valid', () => {
            const { container } = render(
                <BaseControlInput
                    id="test"
                    type={type}
                    label="Hello world"
                    message="Additional message"
                    isValid={true}
                />
            )
            expect(container.firstChild).toMatchSnapshot()
        })
    })
})
