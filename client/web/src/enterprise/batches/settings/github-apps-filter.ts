import {BatchChangesCodeHostFields, ExternalServiceKind} from '../../../graphql-operations'

// This function helps us figure out if an installation is in progress, until we have a more permanent state
// via SRCH-798. Once that issues is closed, this function can maybe be dropped.
export function credentialForGitHubAppExists(
    appName: string | null,
    nodes: BatchChangesCodeHostFields[] | undefined
): boolean {
    if (!appName) {
        return false;
    }

    if (!nodes || nodes.length === 0) {
        return false;
    }

    return nodes.filter(n =>
        n.externalServiceKind === ExternalServiceKind.GITHUB &&
        n.credential?.gitHubApp?.name === appName)
        .length > 0
}
