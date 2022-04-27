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
              class="container d-flex"
            >
              <input
                class="input form-control with-invalid-icon"
                placeholder="loading status input"
                title="Input loading"
                type="text"
                value="Input value"
              />
              <div
                class="loadingSpinner spinner"
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

        expect(container.firstChild).toMatchInlineSnapshot(`
            <label
              class="label w-100"
            >
              <div
                class="mb-2"
              >
                Input label
              </div>
              <div
                class="container d-flex"
              >
                <input
                  class="input form-control with-invalid-icon"
                  placeholder="loading status input"
                  title="Input loading"
                  type="text"
                  value="Input value"
                />
                <div
                  class="loadingSpinner spinner"
                />
              </div>
              <small
                class="text-muted form-text font-weight-normal mt-2"
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
