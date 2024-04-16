import { render, cleanup, type RenderResult, fireEvent, act } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import sinon from 'sinon'
import { afterAll, beforeEach, describe, expect, test } from 'vitest'

import type { TourTaskStepType, TourTaskType } from '@sourcegraph/shared/src/settings/temporary'
import { MockTemporarySettings } from '@sourcegraph/shared/src/settings/temporary/testUtils'
import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { Tour } from './Tour'

const tourId = 'MockTour'
const StepVideo: TourTaskStepType = {
    id: 'Video',
    label: 'Video',
    action: {
        type: 'video',
        value: '#',
    },
}
const StepLink: TourTaskStepType = {
    id: 'GeneralLink',
    label: 'General Link',
    action: {
        type: 'link',
        value: '#',
    },
}
const StepRestart = {
    id: 'Restart',
    label: 'Restart',
    action: {
        type: 'restart',
        value: 'Restart button title',
    },
} satisfies TourTaskStepType
const mockedTasks: TourTaskType[] = [
    {
        title: 'Task 1',
        steps: [StepLink, StepVideo, StepRestart],
    },
]

const mockedTelemetryService = { ...NOOP_TELEMETRY_SERVICE, log: sinon.spy() }
const setup = (overrideTasks?: TourTaskType[]): RenderResult =>
    render(
        <MemoryRouter initialEntries={['/']}>
            <MockTemporarySettings settings={{}}>
                <Tour
                    telemetryService={mockedTelemetryService}
                    id={tourId}
                    tasks={overrideTasks ?? mockedTasks}
                    defaultSnippets={{}}
                />
            </MockTemporarySettings>
        </MemoryRouter>
    )

describe('Tour.tsx', () => {
    afterAll(cleanup)

    beforeEach(() => {
        mockedTelemetryService.log.resetHistory()
    })

    test('renders and triggers initial event log', () => {
        const { getByTestId } = setup()
        expect(getByTestId('tour-content')).toBeInTheDocument()
        expect(mockedTelemetryService.log.withArgs('TourShown', { tourId }).calledOnce).toBeTruthy()
    })

    test('handles closing tour and triggers event log', () => {
        const { getByTestId } = setup()
        expect(getByTestId('tour-content')).toBeInTheDocument()

        fireEvent.click(getByTestId('tour-close-btn'))

        expect(() => getByTestId('tour-content')).toThrow()
        expect(mockedTelemetryService.log.withArgs('TourClosed', { tourId }).calledOnce).toBeTruthy()
    })

    test('handles "type=video" step and triggers event log', () => {
        const { getByTestId, getByText } = setup()
        // clicking video step will open a video modal
        fireEvent.click(getByText(StepVideo.label))
        expect(getByTestId('modal-video')).toBeInTheDocument()

        // click somewhere outside to close video modal
        fireEvent.click(getByTestId('modal-video-close'))
        expect(
            mockedTelemetryService.log.withArgs('TourStepClicked', { tourId, stepId: StepVideo.id }).calledOnce
        ).toBeTruthy()
    })

    test('handles "type=link" step and triggers event log', () => {
        const { getByText } = setup()
        fireEvent.click(getByText(StepLink.label))
        expect(
            mockedTelemetryService.log.withArgs('TourStepClicked', { tourId, stepId: StepLink.id }).calledOnce
        ).toBeTruthy()
    })

    test('handles "type=restart" and triggers event log', () => {
        const { getByText } = setup()

        fireEvent.click(getByText(StepRestart.action.value))

        expect(mockedTelemetryService.log.withArgs('TourRestartClicked', { tourId }).callCount).toBeTruthy()
    })

    test('handles completing tour and triggers event log', () => {
        const { getByText } = setup([{ title: 'task', steps: [StepLink] }])
        act(() => {
            fireEvent.click(getByText(StepLink.label))
        })
        expect(mockedTelemetryService.log.withArgs('TourCompleted', { tourId }).calledOnce).toBeTruthy()
    })
})
