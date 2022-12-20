import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'
import { renderWithBrandedContext } from '@sourcegraph/wildcard'

import { StatusMessagesNavItem } from './StatusMessagesNavItem'
import { allStatusMessages, newStatusMessageMock } from './StatusMessagesNavItem.mocks'

describe('StatusMessagesNavItem', () => {
    test('no messages', () => {
        expect(
            renderWithBrandedContext(
                <MockedTestProvider mocks={[newStatusMessageMock([])]}>
                    <StatusMessagesNavItem disablePolling={true} />
                </MockedTestProvider>
            ).asFragment()
        ).toMatchSnapshot()
    })

    test('all messages', () => {
        expect(
            renderWithBrandedContext(
                <MockedTestProvider mocks={[newStatusMessageMock(allStatusMessages)]}>
                    <StatusMessagesNavItem disablePolling={true} />
                </MockedTestProvider>
            ).asFragment()
        ).toMatchSnapshot()
    })
})
