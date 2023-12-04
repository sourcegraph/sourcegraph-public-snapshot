import { render } from '@testing-library/react'
import { describe, expect, it } from 'vitest'

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
