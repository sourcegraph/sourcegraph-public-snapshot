export const sanitizeIndexer = (imageName: string): string => {
    const sgPrefix = 'sourcegraph/'
    const [base] = imageName.split('@')
    return base.startsWith(sgPrefix) ? base.substring(sgPrefix.length) : base
}
