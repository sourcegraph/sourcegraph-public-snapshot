import { State, Action } from './context'

export const initialState = (initialStep: number, totalSteps: number): State => ({
    current: initialStep,
    initialStep,
    totalSteps,
    steps: {
        [initialStep]: {
            index: initialStep,
            isFirstStep: true,
            isLastStep: false,
            isVisited: true,
            isComplete: false,
        },
    },
})

export const reducer = (state: State, action: Action): State => {
    switch (action.type) {
        case 'SET_CURRENT_STEP':
            return {
                ...state,
                current: action.payload.index,
                steps: {
                    ...state.steps,
                    [state.steps[action.payload.index].index]: {
                        ...state.steps[action.payload.index],
                        isVisited: true,
                    },
                },
            }
        case 'SET_COMPLETE_STEP':
            return {
                ...state,
                steps: {
                    ...state.steps,
                    [state.steps[action.payload.index].index]: {
                        ...state.steps[action.payload.index],
                        isComplete: action.payload.complete,
                    },
                },
            }
        case 'SET_STEPS':
            return {
                ...state,
                steps: { ...action.payload.steps },
            }
        case 'RESET_TO_THE_RIGHT': {
            for (const currentIndex in state.steps) {
                // skip steps to the left
                if (+currentIndex <= action.payload.index) {
                    continue
                }

                const stepToTheRight = state.steps[currentIndex]
                stepToTheRight.isComplete = false
                stepToTheRight.isVisited = false
            }

            return {
                ...state,
                steps: state.steps,
            }
        }

        default:
            throw new Error('wrong action type')
    }
}
