import React from 'react'
import { CloseDeleteCampaignPrompt } from './CloseDeleteCampaignPrompt'
import { shallow } from 'enzyme'

describe('CloseDeleteCampaignPrompt', () => {
    test('some still open', () =>
        expect(
            shallow(
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
            shallow(
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
