import { TabsApi } from './useTabs'

interface Tab {
    mounted: boolean
    index: number
}

export interface Tabs {
    [key: number]: Tab
}

export interface State extends TabsApi {
    current: number
    tabs: Tabs
}

export type Action =
    | { type: 'MOUNTED_TAB'; payload: { index: number; mounted: boolean } }
    | { type: 'SET_TABS'; payload: { tabs: Tabs } }

export const reducer = (state: State, action: Action): State => {
    switch (action.type) {
        case 'MOUNTED_TAB':
            return {
                ...state,
                current: action.payload.index,
                tabs: {
                    ...state.tabs,
                    [state.tabs[action.payload.index].index]: {
                        ...state.tabs[action.payload.index],
                        mounted: action.payload.mounted,
                    },
                },
            }
        case 'SET_TABS':
            return {
                ...state,
                tabs: { ...action.payload.tabs },
            }
        default:
            throw new Error('wrong action type')
    }
}
