import { render, cleanup, RenderResult, fireEvent, act } from '@testing-library/react'
import { renderHook, RenderHookResult } from '@testing-library/react-hooks'
import { MemoryRouter } from 'react-router'
import sinon from 'sinon'

import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { QuickStartTourListState, useQuickStartTourListState } from '../../../stores/quickStartTourState'

import { Tour } from './Tour'
import { TourLanguage, TourTaskStepType, TourTaskType } from './types'

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
const MockTasks: TourTaskType[] = [
    {
        title: 'Task 1',
        steps: [StepLink, StepVideo, StepLanguageSpecificLink, StepRestart],
    },
]

const mockLog = sinon.spy()
const renderTour = (
    overrideTasks?: TourTaskType[]
): [RenderResult, RenderHookResult<unknown, QuickStartTourListState>] => {
    const componentResult = render(
        <MemoryRouter initialEntries={['/']}>
            <Tour
                telemetryService={{ ...NOOP_TELEMETRY_SERVICE, log: mockLog }}
                id={TourId}
                tasks={overrideTasks ?? MockTasks}
            />
        </MemoryRouter>
    )

    const hookStateResult = renderHook(() => useQuickStartTourListState())

    return [componentResult, hookStateResult]
}

describe('Tour.tsx', () => {
    afterAll(cleanup)

    beforeEach(() => {
        localStorage.clear()
        mockLog.resetHistory()
    })

    test('renders and triggers initial event log', () => {
        const [{ getByTestId }] = renderTour()
        expect(getByTestId('tour-content')).toBeTruthy()
        expect(
            mockLog.withArgs(TourId + 'Shown', { language: undefined }, { language: undefined }).calledOnce
        ).toBeTruthy()
    })

    test('handles closing tour and triggers event log', () => {
        const [{ getByTestId }, { result }] = renderTour()
        expect(getByTestId('tour-content')).toBeTruthy()

        fireEvent.click(getByTestId('tour-close-btn'))

        expect(() => getByTestId('tour-content')).toThrow()
        expect(
            mockLog.withArgs(TourId + 'Closed', { language: undefined }, { language: undefined }).calledOnce
        ).toBeTruthy()
        expect(result.current.tours[TourId].status).toBe('closed')
    })

    test('handles "type=video" step and triggers event log', () => {
        const [{ getByTestId, getByText }, { result }] = renderTour()
        // clicking video step will open a video modal
        fireEvent.click(getByText(StepVideo.label))
        expect(getByTestId('modal-video')).toBeTruthy()

        // click somewhere outside to close video modal
        fireEvent.click(getByTestId('modal-video-close'))
        expect(
            mockLog.withArgs(TourId + StepVideo.id + 'Clicked', { language: undefined }, { language: undefined })
                .calledOnce
        ).toBeTruthy()

        expect(result.current.tours[TourId].completedStepIds?.includes(StepVideo.id)).toBeTruthy()
    })

    test('handles "type=link" step and triggers event log', () => {
        const [{ getByText }, { result }] = renderTour()
        fireEvent.click(getByText(StepLink.label))
        expect(
            mockLog.withArgs(TourId + StepLink.id + 'Clicked', { language: undefined }, { language: undefined })
                .calledOnce
        ).toBeTruthy()
        expect(result.current.tours[TourId].completedStepIds?.includes(StepLink.id)).toBeTruthy()
    })

    test('handles "type=link" language specific step and triggers event log', () => {
        const [{ getByText }, { result }] = renderTour()
        fireEvent.click(getByText(StepLanguageSpecificLink.label))
        expect(
            mockLog.withArgs(
                TourId + StepLanguageSpecificLink.id + 'Clicked',
                { language: undefined },
                { language: undefined }
            ).calledOnce
        ).toBeTruthy()
        fireEvent.click(getByText(TourLanguage.Javascript))

        expect(
            mockLog.withArgs(
                TourId + 'LanguageClicked',
                { language: TourLanguage.Javascript },
                { language: TourLanguage.Javascript }
            ).calledOnce
        ).toBeTruthy()
        expect(
            mockLog.withArgs(
                TourId + StepLanguageSpecificLink.id + 'Clicked',
                { language: TourLanguage.Javascript },
                { language: TourLanguage.Javascript }
            ).calledOnce
        ).toBeTruthy()
        expect(result.current.tours[TourId].completedStepIds?.includes(StepLanguageSpecificLink.id)).toBeTruthy()
    })

    test('handles "type=restart" and triggers event log', () => {
        const [{ getByText }, { result }] = renderTour()

        fireEvent.click(getByText(StepLink.label))
        expect(result.current.tours[TourId].completedStepIds?.includes(StepLink.id)).toBeTruthy()

        fireEvent.click(getByText(StepRestart.action.value as string))

        expect(result.current.tours[TourId]).toEqual({})
        expect(
            mockLog.withArgs(TourId + StepRestart.id + 'Clicked', { language: undefined }, { language: undefined })
                .callCount
        ).toBeTruthy()
    })

    test('handles completing tour and triggers event log', () => {
        const [{ getByText }, { result }] = renderTour([{ title: 'task', steps: [StepLink] }])
        act(() => {
            fireEvent.click(getByText(StepLink.label))
        })
        expect(
            mockLog.withArgs(TourId + 'Completed', { language: undefined }, { language: undefined }).calledOnce
        ).toBeTruthy()
        expect(result.current.tours[TourId].status).toBe('completed')
    })
})
