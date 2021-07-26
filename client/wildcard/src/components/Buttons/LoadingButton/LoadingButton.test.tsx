import { render } from '@testing-library/react'
import React from 'react'

import { LoadingButton } from './LoadingButton'

describe('LoadingButton', () => {
    it('only renders children when loading is false', () => {
        const { container } = render(<LoadingButton loading={false}>Hello world</LoadingButton>)
        expect(container.firstChild).toMatchInlineSnapshot(`
            <button
              class="btn"
              type="button"
            >
              Hello world
            </button>
        `)
    })

    it('only renders loading spinner when loading is true', () => {
        const { container } = render(<LoadingButton loading={true}>Hello world</LoadingButton>)
        expect(container.firstChild).toMatchInlineSnapshot(`
            <button
              class="btn"
              type="button"
            >
              <div
                class="loading-spinner icon-inline"
              />
               
            </button>
        `)
    })

    it('renders both loading spinner and children when loading is true and alwaysShowChildren is set', () => {
        const { container } = render(
            <LoadingButton loading={true} alwaysShowChildren={true}>
                Hello world
            </LoadingButton>
        )
        expect(container.firstChild).toMatchInlineSnapshot(`
            <button
              class="btn"
              type="button"
            >
              <div
                class="loading-spinner icon-inline"
              />
               
              Hello world
            </button>
        `)
    })
})
