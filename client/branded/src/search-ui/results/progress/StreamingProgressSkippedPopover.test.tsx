import { screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import sinon from 'sinon'
import { describe, expect, it } from 'vitest'

import type { Progress } from '@sourcegraph/shared/src/search/stream'
import { noOpTelemetryRecorder } from '@sourcegraph/shared/src/telemetry'
import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { assertAriaDisabled, assertAriaEnabled } from '@sourcegraph/testing'
import { renderWithBrandedContext } from '@sourcegraph/wildcard/src/testing'

import { StreamingProgressSkippedPopover } from './StreamingProgressSkippedPopover'

describe('StreamingProgressSkippedPopover', () => {
    it('should render correctly', () => {
        const progress: Progress = {
            durationMs: 1500,
            matchCount: 2,
            repositoriesCount: 2,
            skipped: [
                {
                    reason: 'repository-fork',
                    message: '10k forked repositories excluded',
                    severity: 'info',
                    title: '10k forked repositories excluded',
                    suggested: {
                        title: 'forked:yes',
                        queryExpression: 'forked:yes',
                    },
                },
                {
                    reason: 'excluded-archive',
                    message: '60k archived repositories excluded',
                    severity: 'info',
                    title: '60k archived repositories excluded',
                    suggested: {
                        title: 'archived:yes',
                        queryExpression: 'archived:yes',
                    },
                },
                {
                    reason: 'excluded-archive',
                    message:
                        'By default we exclude archived repositories. Include them with `archived:yes` in your query.',
                    severity: 'info',
                    title: '1 archived',
                    suggested: {
                        title: 'include archived',
                        queryExpression: 'archived:yes',
                    },
                },
                {
                    reason: 'error',
                    message:
                        'There was a network error retrieving search results. Check your Internet connection and try again.',
                    severity: 'error',
                    title: 'Error loading results',
                },
                {
                    reason: 'shard-timedout',
                    message: 'Search timed out',
                    severity: 'warn',
                    title: 'Search timed out',
                    suggested: {
                        title: 'timeout:2m',
                        queryExpression: 'timeout:2m',
                    },
                },
            ],
        }
        expect(
            renderWithBrandedContext(
                <StreamingProgressSkippedPopover
                    query=""
                    progress={progress}
                    telemetryService={NOOP_TELEMETRY_SERVICE}
                    telemetryRecorder={noOpTelemetryRecorder}
                    onSearchAgain={sinon.spy()}
                />
            ).asFragment()
        ).toMatchSnapshot()
    })

    it('should not show Search Again section if no suggested searches are set', () => {
        const progress: Progress = {
            durationMs: 1500,
            matchCount: 2,
            repositoriesCount: 2,
            skipped: [
                {
                    reason: 'repository-fork',
                    message: '10k forked repositories excluded',
                    severity: 'info',
                    title: '10k forked repositories excluded',
                },
            ],
        }

        renderWithBrandedContext(
            <StreamingProgressSkippedPopover
                query=""
                progress={progress}
                telemetryService={NOOP_TELEMETRY_SERVICE}
                telemetryRecorder={noOpTelemetryRecorder}
                onSearchAgain={sinon.spy()}
            />
        )
        expect(screen.queryByTestId('popover-form')).not.toBeInTheDocument()
    })

    it('should have Search Again button disabled by default', () => {
        const progress: Progress = {
            durationMs: 1500,
            matchCount: 2,
            repositoriesCount: 2,
            skipped: [
                {
                    reason: 'repository-fork',
                    message: '10k forked repositories excluded',
                    severity: 'info',
                    title: '10k forked repositories excluded',
                    suggested: {
                        title: 'forked:yes',
                        queryExpression: 'forked:yes',
                    },
                },
            ],
        }

        renderWithBrandedContext(
            <StreamingProgressSkippedPopover
                query=""
                progress={progress}
                telemetryService={NOOP_TELEMETRY_SERVICE}
                telemetryRecorder={noOpTelemetryRecorder}
                onSearchAgain={sinon.spy()}
            />
        )
        const form = screen.getByTestId('popover-form')
        const searchAgainButton = within(form).getByRole('button')
        expect(searchAgainButton).toBeInTheDocument()
        assertAriaDisabled(searchAgainButton)
    })

    it('should enable Search Again button if at least one item is checked', () => {
        const progress: Progress = {
            durationMs: 1500,
            matchCount: 2,
            repositoriesCount: 2,
            skipped: [
                {
                    reason: 'repository-fork',
                    message: '10k forked repositories excluded',
                    severity: 'info',
                    title: '10k forked repositories excluded',
                    suggested: {
                        title: 'forked:yes',
                        queryExpression: 'forked:yes',
                    },
                },
                {
                    reason: 'excluded-archive',
                    message: '60k archived repositories excluded',
                    severity: 'info',
                    title: '60k archived repositories excluded',
                    suggested: {
                        title: 'archived:yes',
                        queryExpression: 'archived:yes',
                    },
                },
                {
                    reason: 'shard-timedout',
                    message: 'Search timed out',
                    severity: 'warn',
                    title: 'Search timed out',
                    suggested: {
                        title: 'timeout:2m',
                        queryExpression: 'timeout:2m',
                    },
                },
            ],
        }

        renderWithBrandedContext(
            <StreamingProgressSkippedPopover
                query=""
                progress={progress}
                telemetryService={NOOP_TELEMETRY_SERVICE}
                telemetryRecorder={noOpTelemetryRecorder}
                onSearchAgain={sinon.spy()}
            />
        )

        const checkboxes = screen.getAllByTestId(/^streaming-progress-skipped-suggest-check/)
        expect(checkboxes).toHaveLength(3)
        userEvent.click(checkboxes[1])

        const form = screen.getByTestId('popover-form')
        const searchAgainButton = within(form).getByRole('button')
        expect(searchAgainButton).toBeInTheDocument()
        assertAriaEnabled(searchAgainButton)
    })

    it('should disable Search Again button if unchecking all items', () => {
        const progress: Progress = {
            durationMs: 1500,
            matchCount: 2,
            repositoriesCount: 2,
            skipped: [
                {
                    reason: 'repository-fork',
                    message: '10k forked repositories excluded',
                    severity: 'info',
                    title: '10k forked repositories excluded',
                    suggested: {
                        title: 'forked:yes',
                        queryExpression: 'forked:yes',
                    },
                },
                {
                    reason: 'excluded-archive',
                    message: '60k archived repositories excluded',
                    severity: 'info',
                    title: '60k archived repositories excluded',
                    suggested: {
                        title: 'archived:yes',
                        queryExpression: 'archived:yes',
                    },
                },
                {
                    reason: 'shard-timedout',
                    message: 'Search timed out',
                    severity: 'warn',
                    title: 'Search timed out',
                    suggested: {
                        title: 'timeout:2m',
                        queryExpression: 'timeout:2m',
                    },
                },
            ],
        }

        renderWithBrandedContext(
            <StreamingProgressSkippedPopover
                query=""
                progress={progress}
                telemetryService={NOOP_TELEMETRY_SERVICE}
                telemetryRecorder={noOpTelemetryRecorder}
                onSearchAgain={sinon.spy()}
            />
        )

        const checkboxes = screen.getAllByTestId(/^streaming-progress-skipped-suggest-check/)
        expect(checkboxes).toHaveLength(3)
        userEvent.click(checkboxes[1])

        const form = screen.getByTestId('popover-form')
        const searchAgainButton = within(form).getByRole('button')
        assertAriaEnabled(searchAgainButton)

        userEvent.click(checkboxes[1])
        assertAriaDisabled(searchAgainButton)
    })

    it('should call onSearchAgain with selected items when button is clicked', () => {
        const progress: Progress = {
            durationMs: 1500,
            matchCount: 2,
            repositoriesCount: 2,
            skipped: [
                {
                    reason: 'shard-timedout',
                    message: 'Search timed out',
                    severity: 'warn',
                    title: 'Search timed out',
                    suggested: {
                        title: 'timeout:2m',
                        queryExpression: 'timeout:2m',
                    },
                },
                {
                    reason: 'repository-fork',
                    message: '10k forked repositories excluded',
                    severity: 'info',
                    title: '10k forked repositories excluded',
                    suggested: {
                        title: 'forked:yes',
                        queryExpression: 'forked:yes',
                    },
                },
                {
                    reason: 'excluded-archive',
                    message: '60k archived repositories excluded',
                    severity: 'info',
                    title: '60k archived repositories excluded',
                    suggested: {
                        title: 'archived:yes',
                        queryExpression: 'archived:yes',
                    },
                },
            ],
        }

        const searchAgain = sinon.spy()

        renderWithBrandedContext(
            <StreamingProgressSkippedPopover
                query=""
                progress={progress}
                telemetryService={NOOP_TELEMETRY_SERVICE}
                telemetryRecorder={noOpTelemetryRecorder}
                onSearchAgain={searchAgain}
            />
        )

        const checkboxes = screen.getAllByTestId(/^streaming-progress-skipped-suggest-check/)
        expect(checkboxes).toHaveLength(3)
        const checkbox1 = checkboxes[1]
        userEvent.click(checkbox1)

        expect(checkboxes).toHaveLength(3)
        const checkbox2 = checkboxes[2]
        userEvent.click(checkbox2)

        const form = screen.getByTestId('popover-form')
        const submitButton = within(form).getByRole('button')
        userEvent.click(submitButton)

        sinon.assert.calledOnce(searchAgain)
        sinon.assert.calledWith(searchAgain, ['forked:yes', 'archived:yes'])
    })
})
