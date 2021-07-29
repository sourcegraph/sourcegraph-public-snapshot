import { render } from '@testing-library/react'
import React from 'react'

import { Select } from './Select'

describe('Select', () => {
    it('renders correctly', () => {
        const { container } = render(
            <Select label="What is your favorite fruit?" message="Hello world">
                <option value="">Select a value</option>
                <option value="apples">Apples</option>
                <option value="bananas">Bananas</option>
                <option value="oranges">Oranges</option>
            </Select>
        )

        expect(container.firstChild).toMatchInlineSnapshot(`
            <div
              class="form-check"
            >
              <label
                class="form-check-label"
              >
                <select
                  class="form-control"
                >
                  <option
                    value=""
                  >
                    Select a value
                  </option>
                  <option
                    value="apples"
                  >
                    Apples
                  </option>
                  <option
                    value="bananas"
                  >
                    Bananas
                  </option>
                  <option
                    value="oranges"
                  >
                    Oranges
                  </option>
                </select>
                What is your favorite fruit?
              </label>
              <small
                class="field-message"
              >
                Hello world
              </small>
            </div>
        `)
    })
})
