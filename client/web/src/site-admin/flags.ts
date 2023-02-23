export const isPackagesEnabled = (): boolean =>
    window.context?.experimentalFeatures?.npmPackages === 'enabled' ||
    window.context?.experimentalFeatures?.goPackages === 'enabled' ||
    window.context?.experimentalFeatures?.jvmPackages === 'enabled' ||
    window.context?.experimentalFeatures?.rubyPackages === 'enabled' ||
    window.context?.experimentalFeatures?.pythonPackages === 'enabled' ||
    window.context?.experimentalFeatures?.rustPackages === 'enabled'
