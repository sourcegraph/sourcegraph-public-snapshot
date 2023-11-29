/**
 * Returns true if the given service type is a package host.
 *
 * Ideally, the backend would tell us which repos are packages. This function is
 * a temporary workaround until that is implemented. There are already many
 * different locations we have to manually update when adding a new package
 * host, and this function is just one of those places.
 */
export function isPackageServiceType(serviceType?: string): boolean {
    switch (serviceType) {
        case 'jvmPackages':
        case 'npmPackages':
        case 'pythonPackages':
        case 'rubyPackages':
        case 'goModules':
        case 'rustPackages': {
            return true
        }
        default: {
            return false
        }
    }
}
