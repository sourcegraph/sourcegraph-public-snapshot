export type SetupChecklistItem = {
    name: string
    description: React.ReactNode
    setupURL: string
    docsURL: string
    configured?: boolean
}

const AddLicenseStep: SetupChecklistItem = {
    name: 'Add License Key',
    setupURL: '/setup/add-license',
    docsURL: '/todo:',
    description: (
        <>
            Sourcegraph uses an SMTP server of your choosing to send emails for
            <ul>
                <li>code monitoring notification</li>
                <li>important updates to user accounts</li>
                <li>built-in authentication</li>
                <li>inviting other users to the instance</li>
            </ul>
        </>
    ),
}

const CodeHostStep: SetupChecklistItem = {
    name: 'Code hosts',
    setupURL: '/setup/remote-repositories',
    docsURL: '/admin/repo/add',
    description: (
        <>
            Sourcegraph uses an SMTP server of your choosing to send emails for
            <ul>
                <li>code monitoring notification</li>
                <li>important updates to user accounts</li>
                <li>built-in authentication</li>
                <li>inviting other users to the instance</li>
            </ul>
        </>
    ),
}

export function useSetupChecklist() {
    return {
        data: [AddLicenseStep, CodeHostStep],
        loading: false,
    }
}
