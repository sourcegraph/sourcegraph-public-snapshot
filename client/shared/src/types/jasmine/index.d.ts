interface CustomReporterResult {
    status: 'passed' | 'failed' | 'disabled' | 'pending'
    fullName: string
}
