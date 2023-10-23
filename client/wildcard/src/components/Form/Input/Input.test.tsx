import { describe, expect, it } from '@jest/globals'
import { render } from '@testing-library/react'

import { Input } from './Input'

const STATUS = ['loading', 'error', 'valid'] as const

describe('Input', () => {
    it('renders an input correctly', () => {
        const { container } = render(
            <Input
                defaultValue="Input value"
                title="Input loading"
                message="random message"
                status="loading"
                placeholder="loading status input"
            />
        )

        expect(container.firstChild).toMatchInlineSnapshot(`
            <div
              class="container loader-input loaderInput"
            >
              <input
                class="inputLoading form-control with-invalid-icon"
                placeholder="loading status input"
                title="Input loading"
                type="text"
                value="Input value"
              />
              <div
                aria-label="Loading"
                aria-live="polite"
                class="mdi-icon loadingSpinner spinner"
                data-loading-spinner="true"
                role="img"
              />
            </div>
        `)
    })

    it('renders an input with label correctly', () => {
        const { container } = render(
            <Input
                defaultValue="Input value"
                title="Input loading"
                message="random message"
                status="loading"
                placeholder="loading status input"
                label="Input label"
            />
        )

        expect(container.firstChild).toMatchSnapshot()
    })

    it.each(STATUS)("Renders status '%s' correctly", status => {
        const { container } = render(<Input status={status} defaultValue="" />)
        expect(container.firstChild).toMatchSnapshot()
    })
})
