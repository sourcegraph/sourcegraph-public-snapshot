export interface SetupChecklistItem {
    name: string
    setupURL: string
    configured?: boolean
}

const AddLicenseStep: SetupChecklistItem = {
    name: 'Add License Key',
    setupURL: '/setup/add-license',
}

const CodeHostStep: SetupChecklistItem = {
    name: 'Code hosts',
    setupURL: '/setup/remote-repositories',
}

interface UseSetupChecklistReturnType {
    data: SetupChecklistItem[]
    error?: any
    loading: boolean
}

export function useSetupChecklist(): UseSetupChecklistReturnType {
    return {
        data: [AddLicenseStep, CodeHostStep],
        loading: false,
    }
}
