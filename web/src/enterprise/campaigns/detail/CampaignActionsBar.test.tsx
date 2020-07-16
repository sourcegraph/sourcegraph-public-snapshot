import React from 'react'
import { CampaignActionsBar } from './CampaignActionsBar'
import { shallow } from 'enzyme'

const PROPS = {
    name: 'Super campaign',
    formID: 'form1',
    onNameChange: () => undefined,
    onEdit: () => undefined,
    // eslint-disable-next-line @typescript-eslint/require-await
    onClose: async () => undefined,
    // eslint-disable-next-line @typescript-eslint/require-await
    onDelete: async () => undefined,
}

describe('CampaignActionsBar', () => {
    test('new with patch set', () =>
        expect(shallow(<CampaignActionsBar {...PROPS} mode="viewing" campaign={undefined} />)).toMatchSnapshot())
    test('new without patch set', () =>
        expect(shallow(<CampaignActionsBar {...PROPS} mode="viewing" campaign={undefined} />)).toMatchSnapshot())
    test('not editable', () =>
        expect(
            shallow(
                <CampaignActionsBar
                    {...PROPS}
                    mode="viewing"
                    campaign={{
                        closedAt: null,
                        name: 'Super campaign',
                        viewerCanAdminister: false,
                    }}
                />
            )
        ).toMatchSnapshot())
    test('editable', () =>
        expect(
            shallow(
                <CampaignActionsBar
                    {...PROPS}
                    mode="viewing"
                    campaign={{
                        closedAt: null,
                        name: 'Super campaign',
                        viewerCanAdminister: true,
                    }}
                />
            )
        ).toMatchSnapshot())
    test('closed', () =>
        expect(
            shallow(
                <CampaignActionsBar
                    {...PROPS}
                    mode="viewing"
                    campaign={{
                        closedAt: new Date().toISOString(),
                        name: 'Super campaign',
                        viewerCanAdminister: true,
                    }}
                />
            )
        ).toMatchSnapshot())
    test('edit mode', () =>
        expect(
            shallow(
                <CampaignActionsBar
                    {...PROPS}
                    mode="editing"
                    campaign={{
                        closedAt: null,
                        name: 'Super campaign',
                        viewerCanAdminister: true,
                    }}
                />
            )
        ).toMatchSnapshot())
    test('processing', () =>
        expect(
            shallow(
                <CampaignActionsBar
                    {...PROPS}
                    mode="editing"
                    campaign={{
                        closedAt: null,
                        name: 'Super campaign',
                        viewerCanAdminister: true,
                    }}
                />
            )
        ).toMatchSnapshot())
    test('mode: saving', () =>
        expect(
            shallow(
                <CampaignActionsBar
                    {...PROPS}
                    mode="saving"
                    campaign={{
                        closedAt: null,
                        name: 'Super campaign',
                        viewerCanAdminister: true,
                    }}
                />
            )
        ).toMatchSnapshot())
    test('mode: deleting', () =>
        expect(
            shallow(
                <CampaignActionsBar
                    {...PROPS}
                    mode="deleting"
                    campaign={{
                        closedAt: null,
                        name: 'Super campaign',
                        viewerCanAdminister: true,
                    }}
                />
            )
        ).toMatchSnapshot())
    test('mode: closing', () =>
        expect(
            shallow(
                <CampaignActionsBar
                    {...PROPS}
                    mode="closing"
                    campaign={{
                        closedAt: null,
                        name: 'Super campaign',
                        viewerCanAdminister: true,
                    }}
                />
            )
        ).toMatchSnapshot())
    test('some changesets still open', () =>
        expect(
            shallow(
                <CampaignActionsBar
                    {...PROPS}
                    mode="viewing"
                    campaign={{
                        closedAt: null,
                        name: 'Super campaign',
                        viewerCanAdminister: true,
                    }}
                />
            )
        ).toMatchSnapshot())
    test('all changesets not open', () =>
        expect(
            shallow(
                <CampaignActionsBar
                    {...PROPS}
                    mode="viewing"
                    campaign={{
                        closedAt: null,
                        name: 'Super campaign',
                        viewerCanAdminister: true,
                    }}
                />
            )
        ).toMatchSnapshot())
})
