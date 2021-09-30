export interface Routes {
    [key: string]: {
        index: string
        description: string
        lineNumber: string
    }
}

export type Action =
    | { type: 'GET_LIST' }
    | { type: 'SET_CURRENT_ROUTE'; payload: { index: string } }
    | { type: 'ADD_ROUTE'; payload: { index: string; description: string; lineNumber: string } }
    | { type: 'REMOVE_ROUTE'; payload: { routes: Routes } }
    | { type: 'UPDATE_DESCRIPTION'; payload: { index: string; description: string } }

export interface State {
    current: string
    routes: Routes
}

export const initialState = ({ current, routes }: { current: string; routes: Routes }): State => ({
    current,
    routes,
})

export const reducer = (state: State, action: Action): State => {
    switch (action.type) {
        case 'GET_LIST':
            return { ...state }
        case 'SET_CURRENT_ROUTE':
            return {
                ...state,
                current: action.payload.index,
            }
        case 'ADD_ROUTE':
            return {
                ...state,
                routes: {
                    ...state.routes,
                    [action.payload.index]: action.payload,
                },
            }
        case 'REMOVE_ROUTE':
            return {
                ...state,
                routes: { ...action.payload.routes },
            }
        case 'UPDATE_DESCRIPTION':
            return {
                ...state,
                routes: {
                    ...state.routes,
                    [state.routes[action.payload.index].index]: {
                        ...state.routes[action.payload.index],
                        description: action.payload.description,
                    },
                },
            }
    }
}
