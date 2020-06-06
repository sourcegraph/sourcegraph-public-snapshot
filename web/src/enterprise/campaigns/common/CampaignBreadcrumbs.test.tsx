import React from 'react'
import { createRenderer } from 'react-test-renderer/shallow'
import { CampaignBreadcrumbs } from './CampaignBreadcrumbs'

describe('CampaignBreadcrumbs', () => {
    test('new', () => expect(createRenderer().render(<CampaignBreadcrumbs />)).toMatchSnapshot())

    test('existing', () =>
        expect(
            createRenderer().render(
                <CampaignBreadcrumbs campaign={{ name: 'My campaign', url: 'https://example.com' }} />
            )
        ).toMatchSnapshot())
})
