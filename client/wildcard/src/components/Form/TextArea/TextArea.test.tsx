import { render } from '@testing-library/react'
import React from 'react'

import { TextArea } from './TextArea'

const ERROR = [true, false] as const

describe('TextArea', () => {
    it('renders an TextArea correctly', () => {
        const { container } = render(
            <TextArea title="TextArea loading" message="random message" placeholder="TextArea" />
        )
        expect(container.firstChild).toMatchInlineSnapshot(`
            <label
              class="w-100"
            >
              <textarea
                class="textarea form-control"
                placeholder="TextArea"
                title="TextArea loading"
              />
              <small
                class="text-muted form-text"
              >
                random message
              </small>
            </label>
        `)
    })

    it.each(ERROR)("Renders status '%s' correctly", status => {
        const { container } = render(<TextArea isError={status} defaultValue="" />)
        expect(container.firstChild).toMatchSnapshot()
    })
})
