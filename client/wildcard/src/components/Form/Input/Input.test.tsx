import { render } from '@testing-library/react'

import { Input } from './Input'

const INPUT_STATUS = ['error', 'valid'] as const

describe('Input', () => {
    describe('Input - does not support loading state', () => {
        it('renders an input correctly', () => {
            const { container } = render(
                <Input
                    defaultValue="Input value"
                    title="Input"
                    label="Input label"
                    message="random message"
                    status="initial"
                    placeholder="initial status input"
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
                  <input
                    class="form-control with-invalid-icon"
                    placeholder="initial status input"
                    title="Input"
                    type="text"
                    value="Input value"
                  />
                  <small
                    class="text-muted form-text font-weight-normal mt-2"
                  >
                    random message
                  </small>
                </label>
            `)
        })

        it.each(INPUT_STATUS)("Renders Input status '%s' correctly", status => {
            const { container } = render(<Input status={status} defaultValue="" />)
            expect(container.firstChild).toMatchSnapshot()
        })
    })
    describe('FormInput - supports loading state', () => {
        it('renders an input correctly', () => {
            const { container } = render(
                <Input
                    defaultValue="Input value"
                    title="Valid input"
                    message="random message"
                    status="valid"
                    placeholder="loading status input"
                />
            )

            expect(container.firstChild).toMatchInlineSnapshot(`
                <input
                  class="form-control with-invalid-icon is-valid"
                  placeholder="loading status input"
                  title="Valid input"
                  type="text"
                  value="Input value"
                />
            `)
        })

        it('renders an input with label correctly', () => {
            const { container } = render(
                <Input
                    defaultValue="Input value"
                    title="Error input"
                    message="random message"
                    status="error"
                    placeholder="error status input"
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
                  <input
                    class="form-control with-invalid-icon is-invalid"
                    placeholder="error status input"
                    title="Error input"
                    type="text"
                    value="Input value"
                  />
                  <small
                    class="text-muted form-text font-weight-normal mt-2"
                  >
                    random message
                  </small>
                </label>
            `)
        })

        it.each(INPUT_STATUS)("Renders FormInput status '%s' correctly", status => {
            const { container } = render(<Input status={status} defaultValue="" />)
            expect(container.firstChild).toMatchSnapshot()
        })
    })
})
