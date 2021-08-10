import { render } from '@testing-library/react'
import React from 'react'

import { BaseControlInput, BASE_CONTROL_TYPES } from './BaseControlInput'

describe('BaseControlInput', () => {
    describe.each(BASE_CONTROL_TYPES)('%s input', type => {
        it('renders correctly', () => {
            const { container } = render(<BaseControlInput id="test" type={type} label="Hello world" />)
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
