import { describe, expect, it } from '@jest/globals'
import { render } from '@testing-library/react'

import { FeedbackText } from './FeedbackText'

describe('FeedbackText', () => {
    it('render a FeedbackText', () => {
        const { asFragment } = render(<FeedbackText />)
        expect(asFragment()).toMatchSnapshot()
    })
    it('render feedbackText with header', () => {
        const { asFragment } = render(<FeedbackText headerText="This is a header text" />)
        expect(asFragment()).toMatchSnapshot()
    })
    it('render feedbackText with footer', () => {
        const { asFragment } = render(<FeedbackText footerText="This is a footer text" />)
        expect(asFragment()).toMatchSnapshot()
    })
})
