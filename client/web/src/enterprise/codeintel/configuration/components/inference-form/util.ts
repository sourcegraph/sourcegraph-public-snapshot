export const sanitizeIndexer = (imageName: string): string => {
    const sgPrefix = 'sourcegraph/'
    const [base] = imageName.split('@')
    return base.startsWith(sgPrefix) ? base.slice(sgPrefix.length) : base
}

export const sanitizeRoot = (root: string): string => {
    if (root === '/' || root === '') {
        return '/'
    }

    return `/${root}`
}
