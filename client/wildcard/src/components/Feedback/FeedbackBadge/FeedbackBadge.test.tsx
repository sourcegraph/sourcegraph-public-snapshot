import { describe, expect, it } from '@jest/globals'
import { render } from '@testing-library/react'

import { FEEDBACK_BADGES_STATUS } from './constant'
import { FeedbackBadge } from './FeedbackBadge'

describe('FeedbackBadge', () => {
    it('renders a FeedbackBadge', () => {
        const { container } = render(<FeedbackBadge status="new" feedback={{ mailto: 'support@sourcegraph.com' }} />)
        expect(container.firstChild).toMatchSnapshot()
    })
    it.each(FEEDBACK_BADGES_STATUS)("Renders status '%s' correctly", status => {
        const { container } = render(<FeedbackBadge status={status} feedback={{ mailto: 'support@sourcegraph.com' }} />)
        expect(container.firstChild).toMatchSnapshot()
    })
})
