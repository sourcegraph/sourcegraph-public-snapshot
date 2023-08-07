export type SetupChecklistItem = {
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

export function useSetupChecklist() {
    return {
        data: [AddLicenseStep, CodeHostStep],
        loading: false,
    }
}
