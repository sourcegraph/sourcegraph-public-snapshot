import { render } from '@testing-library/react'
import { describe, expect, it } from 'vitest'

import { Select, type SelectProps } from './Select'

describe('Select', () => {
    const renderSelect = (selectProps?: Partial<SelectProps>) =>
        render(
            <Select id="test" label="What is your favorite fruit?" {...selectProps}>
                <option value="">Select a value</option>
                <option value="apples">Apples</option>
                <option value="bananas">Bananas</option>
                <option value="oranges">Oranges</option>
            </Select>
        )

    describe.each(['native', 'custom'])('%s variant', type => {
        const isCustomStyle = type === 'custom'

        it('renders correctly', () => {
            const { container } = renderSelect({ isCustomStyle })
            expect(container.firstChild).toMatchSnapshot()
        })

        it('renders with message correctly', () => {
            const { container } = renderSelect({ isCustomStyle, message: 'Additional message' })
            expect(container.firstChild).toMatchSnapshot()
        })

        it('renders with correct styles when invalid', () => {
            const { container } = renderSelect({ isCustomStyle, message: 'Additional message', isValid: false })
            expect(container.firstChild).toMatchSnapshot()
        })

        it('renders with correct styles when valid', () => {
            const { container } = renderSelect({ isCustomStyle, message: 'Additional message', isValid: true })
            expect(container.firstChild).toMatchSnapshot()
        })
    })
})
