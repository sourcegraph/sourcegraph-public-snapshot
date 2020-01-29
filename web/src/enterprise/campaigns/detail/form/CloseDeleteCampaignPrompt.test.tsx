import React from 'react'
import { CloseDeleteCampaignPrompt } from './CloseDeleteCampaignPrompt'
import { createRenderer } from 'react-test-renderer/shallow'

describe('CloseDeleteCampaignPrompt', () => {
    test('renders', () =>
        expect(
            createRenderer().render(
                <CloseDeleteCampaignPrompt
                    summary={<span className="btn btn-secondary dropdown-toggle">Close</span>}
                    message={<p>message</p>}
                    changesetsCount={2}
                    closeChangesets={true}
                    onCloseChangesetsToggle={() => undefined}
                    buttonText="Delete"
                    onButtonClick={() => undefined}
                    buttonClassName="btn-danger"
                    buttonDisabled={false}
                    className="c"
                />
            )
        ).toMatchSnapshot())
})
