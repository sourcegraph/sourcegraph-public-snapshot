import { render, cleanup, RenderResult, fireEvent, act } from '@testing-library/react'
import { MemoryRouter } from 'react-router'
import { CompatRouter } from 'react-router-dom-v5-compat'
import sinon from 'sinon'

import { TourLanguage, TourTaskStepType, TourTaskType } from '@sourcegraph/shared/src/settings/temporary'
import { MockTemporarySettings } from '@sourcegraph/shared/src/settings/temporary/testUtils'
import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { Tour } from './Tour'

const TourId = 'MockTour'
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
const StepRestart: TourTaskStepType = {
    id: 'Restart',
    label: 'Restart',
    action: {
        type: 'restart',
        value: 'Restart button title',
    },
}
const StepLanguageSpecificLink: TourTaskStepType = {
    id: 'LanguageSpecificLink',
    label: 'Language Specific Link',
    action: {
        type: 'link',
        value: {
            [TourLanguage.C]: '#',
            [TourLanguage.Go]: '#',
            [TourLanguage.Java]: '#',
            [TourLanguage.Javascript]: '#',
            [TourLanguage.Php]: '#',
            [TourLanguage.Python]: '#',
            [TourLanguage.Typescript]: '#',
        },
    },
}
const mockedTasks: TourTaskType[] = [
    {
        title: 'Task 1',
        steps: [StepLink, StepVideo, StepLanguageSpecificLink, StepRestart],
    },
]

const mockedTelemetryService = { ...NOOP_TELEMETRY_SERVICE, log: sinon.spy() }
const setup = (overrideTasks?: TourTaskType[]): RenderResult =>
    render(
        <MemoryRouter initialEntries={['/']}>
            <CompatRouter>
                <MockTemporarySettings settings={{}}>
                    <Tour telemetryService={mockedTelemetryService} id={TourId} tasks={overrideTasks ?? mockedTasks} />
                </MockTemporarySettings>
            </CompatRouter>
        </MemoryRouter>
    )

describe('Tour.tsx', () => {
    afterAll(cleanup)

    beforeEach(() => {
        mockedTelemetryService.log.resetHistory()
    })

    test('renders and triggers initial event log', () => {
        const { getByTestId } = setup()
        expect(getByTestId('tour-content')).toBeTruthy()
        expect(
            mockedTelemetryService.log.withArgs(TourId + 'Shown', { language: undefined }, { language: undefined })
                .calledOnce
        ).toBeTruthy()
    })

    test('handles closing tour and triggers event log', () => {
        const { getByTestId } = setup()
        expect(getByTestId('tour-content')).toBeTruthy()

        fireEvent.click(getByTestId('tour-close-btn'))

        expect(() => getByTestId('tour-content')).toThrow()
        expect(
            mockedTelemetryService.log.withArgs(TourId + 'Closed', { language: undefined }, { language: undefined })
                .calledOnce
        ).toBeTruthy()
    })

    test('handles "type=video" step and triggers event log', () => {
        const { getByTestId, getByText } = setup()
        // clicking video step will open a video modal
        fireEvent.click(getByText(StepVideo.label))
        expect(getByTestId('modal-video')).toBeTruthy()

        // click somewhere outside to close video modal
        fireEvent.click(getByTestId('modal-video-close'))
        expect(
            mockedTelemetryService.log.withArgs(
                TourId + StepVideo.id + 'Clicked',
                { language: undefined },
                { language: undefined }
            ).calledOnce
        ).toBeTruthy()
    })

    test('handles "type=link" step and triggers event log', () => {
        const { getByText } = setup()
        fireEvent.click(getByText(StepLink.label))
        expect(
            mockedTelemetryService.log.withArgs(
                TourId + StepLink.id + 'Clicked',
                { language: undefined },
                { language: undefined }
            ).calledOnce
        ).toBeTruthy()
    })

    test('handles "type=link" language specific step and triggers event log', () => {
        const { getByText } = setup()
        fireEvent.click(getByText(StepLanguageSpecificLink.label))
        expect(
            mockedTelemetryService.log.withArgs(
                TourId + StepLanguageSpecificLink.id + 'Clicked',
                { language: undefined },
                { language: undefined }
            ).calledOnce
        ).toBeTruthy()
        fireEvent.click(getByText(TourLanguage.Javascript))

        expect(
            mockedTelemetryService.log.withArgs(
                TourId + 'LanguageClicked',
                { language: TourLanguage.Javascript },
                { language: TourLanguage.Javascript }
            ).calledOnce
        ).toBeTruthy()
        expect(
            mockedTelemetryService.log.withArgs(
                TourId + StepLanguageSpecificLink.id + 'Clicked',
                { language: TourLanguage.Javascript },
                { language: TourLanguage.Javascript }
            ).calledOnce
        ).toBeTruthy()
    })

    test('handles "type=restart" and triggers event log', () => {
        const { getByText } = setup()

        fireEvent.click(getByText(StepRestart.action.value as string))

        expect(
            mockedTelemetryService.log.withArgs(
                TourId + StepRestart.id + 'Clicked',
                { language: undefined },
                { language: undefined }
            ).callCount
        ).toBeTruthy()
    })

    test('handles completing tour and triggers event log', () => {
        const { getByText } = setup([{ title: 'task', steps: [StepLink] }])
        act(() => {
            fireEvent.click(getByText(StepLink.label))
        })
        expect(
            mockedTelemetryService.log.withArgs(TourId + 'Completed', { language: undefined }, { language: undefined })
                .calledOnce
        ).toBeTruthy()
    })
})
