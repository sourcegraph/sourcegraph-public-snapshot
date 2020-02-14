import React from 'react'
import { CloseDeleteCampaignPrompt } from './CloseDeleteCampaignPrompt'
import { createRenderer } from 'react-test-renderer/shallow'

describe('CloseDeleteCampaignPrompt', () => {
    test('renders', () =>
        expect(
            createRenderer().render(
                <CloseDeleteCampaignPrompt
                    disabled={false}
                    disabledTooltip="Cannot delete while campaign is processing..."
                    message={<p>message</p>}
                    changesetsCount={2}
                    buttonText="Delete"
                    onButtonClick={() => undefined}
                    buttonClassName="btn-danger"
                />
            )
        ).toMatchSnapshot())
})
