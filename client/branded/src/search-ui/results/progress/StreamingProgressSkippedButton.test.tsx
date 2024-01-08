import { screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import sinon from 'sinon'
import { beforeAll, describe, expect, it } from 'vitest'

import type { Progress } from '@sourcegraph/shared/src/search/stream'
import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { renderWithBrandedContext } from '@sourcegraph/wildcard/src/testing'

import { StreamingProgressSkippedButton } from './StreamingProgressSkippedButton'

describe('StreamingProgressSkippedButton', () => {
    beforeAll(() => {
        ;(global as any).document.createRange = () => ({
            setStart: () => {},
            setEnd: () => {},
            commonAncestorContainer: {
                nodeName: 'BODY',
                ownerDocument: document,
            },
        })
    })

    it('should be in info state with only info items', () => {
        const progress: Progress = {
            durationMs: 1500,
            matchCount: 2,
            repositoriesCount: 2,
            skipped: [
                {
                    reason: 'excluded-fork',
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

        renderWithBrandedContext(
            <StreamingProgressSkippedButton
                query=""
                progress={progress}
                telemetryService={NOOP_TELEMETRY_SERVICE}
                onSearchAgain={sinon.spy()}
            />
        )
        expect(screen.getByTestId('streaming-progress-skipped')).toBeInTheDocument()
        expect(screen.queryByTestId('streaming-progress-skipped')).not.toHaveClass('outline-danger')
    })

    it('should be in warning state with at least one warning item', () => {
        const progress: Progress = {
            durationMs: 1500,
            matchCount: 2,
            repositoriesCount: 2,
            skipped: [
                {
                    reason: 'excluded-fork',
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
            <StreamingProgressSkippedButton
                query=""
                progress={progress}
                telemetryService={NOOP_TELEMETRY_SERVICE}
                onSearchAgain={sinon.spy()}
            />
        )
        expect(screen.getByTestId('streaming-progress-skipped')).toHaveClass('btnOutline btnDanger')
        expect(screen.queryByTestId('streaming-progress-skipped')).not.toHaveClass('btnOutline btnSecondary')
    })

    it('should open and close popover when button is clicked', async () => {
        const progress: Progress = {
            durationMs: 1500,
            matchCount: 2,
            repositoriesCount: 2,
            skipped: [
                {
                    reason: 'excluded-fork',
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

        renderWithBrandedContext(
            <StreamingProgressSkippedButton
                query=""
                progress={progress}
                telemetryService={NOOP_TELEMETRY_SERVICE}
                onSearchAgain={sinon.spy()}
            />
        )

        const button = screen.getByTestId('streaming-progress-skipped')

        expect(button).toHaveAttribute('aria-expanded', 'false')

        userEvent.click(button)

        await waitFor(() => expect(button).toHaveAttribute('aria-expanded', 'true'))

        userEvent.click(button)

        await waitFor(() => expect(button).toHaveAttribute('aria-expanded', 'false'))
    })

    it('should close popup and call onSearchAgain callback when popover raises event', async () => {
        const progress: Progress = {
            durationMs: 1500,
            matchCount: 2,
            repositoriesCount: 2,
            skipped: [
                {
                    reason: 'excluded-fork',
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

        const onSearchAgain = sinon.spy()

        renderWithBrandedContext(
            <StreamingProgressSkippedButton
                query=""
                progress={progress}
                telemetryService={NOOP_TELEMETRY_SERVICE}
                onSearchAgain={onSearchAgain}
            />
        )
        const toggleButton = screen.getByTestId('streaming-progress-skipped')

        userEvent.click(toggleButton)

        await waitFor(() => {
            // dropdown is opened
            expect(toggleButton).toHaveAttribute('aria-expanded', 'true')
        })

        // Trigger onSearchAgain event and check for changes
        // Find `archived:yes` checkbox
        userEvent.click(screen.getAllByTestId(/^streaming-progress-skipped-suggest-check/)[1], undefined, {
            skipPointerEventsCheck: true,
        })
        userEvent.click(screen.getByTestId('skipped-popover-form-submit-btn'), undefined, {
            skipPointerEventsCheck: true,
        })

        await waitFor(() => {
            // dropdown is closed
            expect(toggleButton).toHaveAttribute('aria-expanded', 'false')
        })

        sinon.assert.calledOnce(onSearchAgain)
        sinon.assert.calledWith(onSearchAgain, ['archived:yes'])
    })
})
