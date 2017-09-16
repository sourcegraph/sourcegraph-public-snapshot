interface PageError {
    StatusCode: number
    StatusText: string
    Error: string
    ErrorID: string
}

interface Window {
    pageError?: PageError
}
