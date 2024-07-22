import { fireEvent, render, screen } from '@testing-library/react'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import { PatternConstrainedInput } from './PatternConstrainedInput'

describe('PatternConstrainedInput', () => {
    const mockOnChange = vi.fn()
    beforeEach(() => {
        vi.clearAllMocks()
    })

    const PATTERN = '[\\w\\-]+'

    it('renders correctly with initial props', () => {
        render(<PatternConstrainedInput value="" pattern={PATTERN} onChange={mockOnChange} />)
        expect(screen.getByRole('textbox')).toBeInTheDocument()
    })

    it('calls onChange with valid input', () => {
        render(<PatternConstrainedInput value="" pattern={PATTERN} onChange={mockOnChange} />)
        const input = screen.getByRole('textbox')
        fireEvent.change(input, { target: { value: 'a-a-a' } })
        expect(mockOnChange).toHaveBeenCalledWith('a-a-a', true)
        expect(input).not.toHaveAttribute('aria-invalid')
    })

    it('calls onChange with invalid input', () => {
        render(<PatternConstrainedInput value="" pattern={PATTERN} onChange={mockOnChange} />)
        const input = screen.getByRole('textbox')
        fireEvent.change(input, { target: { value: 'invalid input!' } })
        expect(mockOnChange).toHaveBeenCalledWith('invalid input!', false)
        expect(input).toHaveAttribute('aria-invalid', 'true')
    })

    it('replaces spaces with hyphens when replaceSpaces is true', () => {
        render(<PatternConstrainedInput value="" pattern={PATTERN} onChange={mockOnChange} replaceSpaces={true} />)
        const input = screen.getByRole('textbox')
        fireEvent.change(input, { target: { value: 'a a a' } })
        expect(mockOnChange).toHaveBeenCalledWith('a-a-a', true)
    })
})
