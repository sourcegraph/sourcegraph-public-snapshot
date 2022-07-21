import {SourcegraphContext} from "../jscontext";

// make sure we always have a minPasswordLen
export const minPasswordLen = (
    window.context.authMinPasswordLength !== undefined && window.context.authMinPasswordLength > 0
) ? window.context.authMinPasswordLength : 12

export function getPasswordPolicy(): any {
    let passwordPolicyReference = window.context.authPasswordPolicy

    if (!passwordPolicyReference) {
        passwordPolicyReference = window.context.experimentalFeatures.passwordPolicy
    }

    return passwordPolicyReference
}

export function validatePassword(
    context: Pick<SourcegraphContext, 'authProviders' | 'sourcegraphDotComMode' | 'experimentalFeatures' |
        'authPasswordPolicy'>,
    password: string
): string | undefined {

    // minPasswordLen always has a value so we do it first
    if (password.length < minPasswordLen) {
        return 'Password must be at least ' + minPasswordLen.toString() + ' characters.'
    }

    let passwordPolicyReference = getPasswordPolicy()

    if (passwordPolicyReference?.enabled) {
        if (
            passwordPolicyReference?.numberOfSpecialCharacters &&
            passwordPolicyReference.numberOfSpecialCharacters > 0
        ) {
            const specialCharacters = /[!"#$%&'()*+,./:;<=>?@[\]^_`{|}~-]/
            // This must be kept in sync with the security.go checks
            const count = (password.match(specialCharacters) || []).length
            if (
                passwordPolicyReference.numberOfSpecialCharacters &&
                count < passwordPolicyReference.numberOfSpecialCharacters
            ) {
                return (
                    'Password must contain ' +
                    passwordPolicyReference.numberOfSpecialCharacters.toString() +
                    ' special character(s).'
                )
            }
        }

        if (
            passwordPolicyReference.requireAtLeastOneNumber
        ) {
            const validRequireAtLeastOneNumber = /\d+/
            if (password.match(validRequireAtLeastOneNumber) === null) {
                return 'Password must contain at least one number.'
            }
        }

        if (
            passwordPolicyReference.requireUpperandLowerCase
        ) {
            const validUseUpperCase = new RegExp('[A-Z]+')
            if (!validUseUpperCase.test(password)) {
                return 'Password must contain at least one uppercase letter.'
            }
        }

        return undefined
    }

    return undefined

}

export function getPasswordRequirements(
    context: Pick<SourcegraphContext, 'authProviders' | 'sourcegraphDotComMode' | 'experimentalFeatures' |
        'authPasswordPolicy'>
): string {
    let passwordPolicyReference = getPasswordPolicy()

    let requirements: string  = 'Your password must include at least ' + minPasswordLen.toString() + ' characters'

    if (passwordPolicyReference && passwordPolicyReference.enabled) {
        console.log('Using enhanced password policy.')

        if (
            passwordPolicyReference.numberOfSpecialCharacters &&
            passwordPolicyReference.numberOfSpecialCharacters > 0
        ) {
            requirements += ', ' + String(passwordPolicyReference.numberOfSpecialCharacters) + ' special characters'
        }
        if (
            passwordPolicyReference.requireAtLeastOneNumber &&
            passwordPolicyReference.requireAtLeastOneNumber
        ) {
            requirements += ', at least one number'
        }
        if (
            passwordPolicyReference.requireUpperandLowerCase &&
            passwordPolicyReference.requireUpperandLowerCase
        ) {
            requirements += ', at least one uppercase letter'
        }

    }

    return requirements
}
