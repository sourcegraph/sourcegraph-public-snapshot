import { describe, expect, it } from '@jest/globals'
import { render } from '@testing-library/react'

import { TextArea } from './TextArea'

describe('TextArea', () => {
    it('should render correctly', () => {
        const { asFragment } = render(
            <TextArea title="TextArea loading" message="random message" placeholder="TextArea" isValid={true} />
        )
        expect(asFragment()).toMatchSnapshot()
    })
})
