import { render } from '@testing-library/react'
import { describe, expect, it } from 'vitest'

import { Text } from './Text'

describe('Text', () => {
    it('should render correctly', () => {
        expect(
            render(
                <Text size="base" weight="medium" alignment="left" mode="single-line">
                    Hello world
                </Text>
            ).asFragment()
        ).toMatchSnapshot()
    })

    it('should render correctly with `as`', () => {
        expect(render(<Text as="div">I am a Text</Text>).asFragment()).toMatchSnapshot()
    })
})
