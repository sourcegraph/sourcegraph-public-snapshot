export function isProbablyTestFile(filename: string): boolean {
    const f = filename.split('.')
    // To account for other test file path structures
    // adjust this regular expression.
    const isTest = /^(test|spec|tests)(\b|_)|(\b|_)(test|spec|tests)$/

    for (const i of f) {
        if (i === 'test') {
            return true
        }
        if (isTest.test(i)) {
            return true
        }
    }
    return false
}
