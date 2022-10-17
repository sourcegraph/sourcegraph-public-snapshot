import { getExtensionVersion } from '../../shared/util/context'

export const checkUrlPermissions = (url: string): Promise<boolean> => {
    const { host, protocol } = new URL(url)

    // Inject content script whenever a new tab was opened with a URL for which we have permission
    return browser.permissions.contains({
        origins: [`${protocol}//${host}/*`],
    })
}

export const IsProductionVersion = !getExtensionVersion().startsWith('0.0.0')
