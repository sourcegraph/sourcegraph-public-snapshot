import { describe, expect, test } from '@jest/globals'
import { render, fireEvent } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { of } from 'rxjs'
import sinon from 'sinon'

import { EMPTY_SETTINGS_CASCADE } from '@sourcegraph/shared/src/settings/settings'

import type { AuthenticatedUser } from '../../auth'
import type { ListCodeMonitors, ListUserCodeMonitorsVariables } from '../../graphql-operations'

import { CodeMonitoringPage } from './CodeMonitoringPage'
import { mockCodeMonitorNodes } from './testing/util'

const additionalProps = {
    authenticatedUser: {
        id: 'foobar',
        username: 'alice',
        emails: [{ email: 'alice@email.test', isPrimary: true, verified: true }],
    } as AuthenticatedUser,
    fetchUserCodeMonitors: ({ id, first, after }: ListUserCodeMonitorsVariables) =>
        of({
            nodes: mockCodeMonitorNodes,
            pageInfo: {
                endCursor: 'foo10',
                hasNextPage: true,
            },
            totalCount: 12,
        }),
    toggleCodeMonitorEnabled: sinon.spy((id: string, enabled: boolean) => of({ id: 'test', enabled: true })),
    settingsCascade: EMPTY_SETTINGS_CASCADE,
    isLightTheme: false,
    isCodyApp: false,
}

const generateMockFetchMonitors =
    (count: number) =>
    ({ id, first, after }: ListUserCodeMonitorsVariables) => {
        const result: ListCodeMonitors = {
            nodes: mockCodeMonitorNodes.slice(0, count),
            pageInfo: {
                endCursor: `foo${count}`,
                hasNextPage: count > 10,
            },
            totalCount: count,
        }
        return of(result)
    }

describe('CodeMonitoringListPage', () => {
    test('Clicking enabled toggle calls toggleCodeMonitorEnabled', () => {
        const component = render(
            <MemoryRouter initialEntries={['/code-monitoring?tab=list']}>
                <CodeMonitoringPage {...additionalProps} fetchUserCodeMonitors={generateMockFetchMonitors(1)} />
            </MemoryRouter>
        )
        const toggle = component.getByTestId('toggle-monitor-enabled')
        fireEvent.click(toggle)
        expect(additionalProps.toggleCodeMonitorEnabled.calledOnce)
    })

    test('Switching tabs from getting started to empty list works', () => {
        const component = render(
            <MemoryRouter initialEntries={['/code-monitoring?tab=getting-started']}>
                <CodeMonitoringPage {...additionalProps} fetchUserCodeMonitors={generateMockFetchMonitors(0)} />
            </MemoryRouter>
        )
        const codeMonitorsButton = component.getByRole('button', { name: 'Code monitors' })
        fireEvent.click(codeMonitorsButton)

        const emptyListMessage = component.getByText(/no code monitors have been created/i)
        expect(emptyListMessage).toBeInTheDocument()
    })

    test('Switching tabs from list to getting started works', () => {
        const component = render(
            <MemoryRouter initialEntries={['/code-monitoring?tab=list']}>
                <CodeMonitoringPage {...additionalProps} fetchUserCodeMonitors={generateMockFetchMonitors(0)} />
            </MemoryRouter>
        )
        const gettingStartedButton = component.getByRole('button', { name: 'Getting started' })
        fireEvent.click(gettingStartedButton)

        const gettingStartedHeader = component.getByText(/proactively monitor/i)
        expect(gettingStartedHeader).toBeInTheDocument()
    })
})
