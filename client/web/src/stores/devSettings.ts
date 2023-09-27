import create from "zustand"

export const useDevSettings = create<{showDialog: boolean}>(() => {
    return {
        showDialog: false,
    }
})

export function toggleDevSettingsDialog(show?: boolean) {
    useDevSettings.setState(state => ({
        showDialog: show ?? !state.showDialog
    }))
}
