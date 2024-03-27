export const isCodyOnlyLicense = (): boolean =>
    Boolean(
        typeof window !== 'undefined' &&
            !window.context.licenseInfo?.features.codeSearch &&
            window.context.licenseInfo?.features.cody
    )

export const isCodeSearchOnlyLicense = (): boolean =>
    Boolean(
        typeof window !== 'undefined' &&
            window.context.licenseInfo?.features.codeSearch &&
            !window.context.licenseInfo?.features.cody
    )

export const isCodeSearchPlusCodyLicense = (): boolean =>
    Boolean(
        typeof window !== 'undefined' &&
            window.context.licenseInfo?.features.codeSearch &&
            window.context.licenseInfo?.features.cody
    )

interface LicenseFeatures {
    isCodeSearchEnabled: boolean
    isCodyEnabled: boolean
}

export const getLicenseFeatures = (): LicenseFeatures => ({
    isCodeSearchEnabled: Boolean(window.context.licenseInfo?.features.codeSearch),
    isCodyEnabled: Boolean(window.context.licenseInfo?.features.cody),
})
