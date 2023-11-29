import { render } from '@testing-library/react'
import { describe, expect, it } from 'vitest'

import { TextArea } from './TextArea'

describe('TextArea', () => {
    it('should render correctly', () => {
        const { asFragment } = render(
            <TextArea title="TextArea loading" message="random message" placeholder="TextArea" isValid={true} />
        )
        expect(asFragment()).toMatchSnapshot()
    })
})
