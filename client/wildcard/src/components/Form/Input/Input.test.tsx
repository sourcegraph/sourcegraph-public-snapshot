import { render } from '@testing-library/react'
import React from 'react'

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
            <label
              class="w-100"
            >
              <div
                class="mb-2"
              >
                Input loading
              </div>
              <div
                class="loader-input__container d-flex"
              >
                <input
                  class="input form-control with-invalid-icon"
                  placeholder="loading status input"
                  type="text"
                  value="Input value"
                />
                <div
                  class="loading-spinner loader-input__spinner"
                />
              </div>
              <small
                class="text-muted form-text message"
              >
                random message
              </small>
            </label>
        `)
    })

    it.each(STATUS)("Renders status '%s' correctly", status => {
        const { container } = render(<Input status={status} defaultValue="" />)
        expect(container.firstChild).toMatchSnapshot()
    })
})
