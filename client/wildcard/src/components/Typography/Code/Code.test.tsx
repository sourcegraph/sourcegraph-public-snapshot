import { describe, expect, it } from '@jest/globals'
import { render } from '@testing-library/react'

import { Code } from './Code'

describe('Code', () => {
    it('renders a simple text code correctly', () => {
        expect(
            render(
                <Code size="small" weight="regular">
                    I am a code
                </Code>
            ).asFragment()
        ).toMatchSnapshot()
    })

    it('supports rendering as different elements', () => {
        expect(
            render(
                <Code as="div" size="small" weight="regular">
                    I am a code
                </Code>
            ).asFragment()
        ).toMatchSnapshot()
    })
})
