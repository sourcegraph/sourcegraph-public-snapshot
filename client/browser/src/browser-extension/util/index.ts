export const checkUrlPermissions = (url: string): Promise<boolean> => {
    const { host, protocol } = new URL(url)

    // Inject content script whenever a new tab was opened with a URL for which we have permission
    return browser.permissions.contains({
        origins: [`${protocol}//${host}/*`],
    })
}
