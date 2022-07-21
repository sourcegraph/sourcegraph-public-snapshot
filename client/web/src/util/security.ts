
export function getPasswordPolicy(): any {
    let passwordPolicyReference = window.context.authPasswordPolicy

    if (!passwordPolicyReference) {
        passwordPolicyReference = window.context.experimentalFeatures.passwordPolicy
    }

    return passwordPolicyReference
}
