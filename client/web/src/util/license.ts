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
