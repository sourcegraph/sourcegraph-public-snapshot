import { render } from '@testing-library/react'

import { FormInput } from './FormInput'

const FORM_INPUT_STATUS = ['loading', 'error', 'valid'] as const
const INPUT_STATUS = ['error', 'valid'] as const

describe('Input', () => {
    describe('Input - does not support loading state', () => {
        it('renders an input correctly', () => {
            const { container } = render(
                <FormInput
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
                  <div
                    class="container"
                  >
                    <input
                      class="input form-control with-invalid-icon"
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
                  </div>
                </label>
            `)
        })

        it.each(INPUT_STATUS)("Renders Input status '%s' correctly", status => {
            const { container } = render(<FormInput status={status} defaultValue="" />)
            expect(container.firstChild).toMatchSnapshot()
        })
    })
    describe('FormInput - supports loading state', () => {
        it('renders an input correctly', () => {
            const { container } = render(
                <FormInput
                    defaultValue="Input value"
                    title="Input loading"
                    message="random message"
                    status="loading"
                    placeholder="loading status input"
                />
            )

            expect(container.firstChild).toMatchInlineSnapshot(`
                <div
                  class="container"
                >
                  <input
                    class="input form-control with-invalid-icon"
                    placeholder="loading status input"
                    title="Input loading"
                    type="text"
                    value="Input value"
                  />
                  <small
                    class="text-muted form-text font-weight-normal mt-2"
                  >
                    random message
                  </small>
                  <div
                    class="loadingSpinner spinner"
                  />
                </div>
            `)
        })

        it('renders an input with label correctly', () => {
            const { container } = render(
                <FormInput
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
                    class="container"
                  >
                    <input
                      class="input form-control with-invalid-icon"
                      placeholder="loading status input"
                      title="Input loading"
                      type="text"
                      value="Input value"
                    />
                    <small
                      class="text-muted form-text font-weight-normal mt-2"
                    >
                      random message
                    </small>
                    <div
                      class="loadingSpinner spinner"
                    />
                  </div>
                </label>
            `)
        })

        it.each(FORM_INPUT_STATUS)("Renders FormInput status '%s' correctly", status => {
            const { container } = render(<FormInput status={status} defaultValue="" />)
            expect(container.firstChild).toMatchSnapshot()
        })
    })
})
