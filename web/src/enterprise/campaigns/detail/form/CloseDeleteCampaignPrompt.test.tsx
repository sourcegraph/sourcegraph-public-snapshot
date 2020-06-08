import React from 'react'
import { CloseDeleteCampaignPrompt } from './CloseDeleteCampaignPrompt'
import { createRenderer } from 'react-test-renderer/shallow'

describe('CloseDeleteCampaignPrompt', () => {
    test('some still open', () =>
        expect(
            createRenderer().render(
                <CloseDeleteCampaignPrompt
                    disabled={false}
                    disabledTooltip="Cannot delete while campaign is processing..."
                    message={<p>message</p>}
                    buttonText="Delete"
                    onButtonClick={() => undefined}
                    buttonClassName="btn-danger"
                />
            )
        ).toMatchSnapshot())
    test('none still open', () =>
        expect(
            createRenderer().render(
                <CloseDeleteCampaignPrompt
                    disabled={false}
                    disabledTooltip="Cannot delete while campaign is processing..."
                    message={<p>message</p>}
                    buttonText="Delete"
                    onButtonClick={() => undefined}
                    buttonClassName="btn-danger"
                />
            )
        ).toMatchSnapshot())
})
