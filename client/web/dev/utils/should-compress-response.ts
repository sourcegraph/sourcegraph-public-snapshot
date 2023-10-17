import compression, { type CompressionFilter } from 'compression'

export const STREAMING_ENDPOINTS = ['/search/stream', '/.api/compute/stream', '/.api/completions/stream']

export const shouldCompressResponse: CompressionFilter = (request, response) => {
    // Disable compression because gzip buffers the full response
    // before sending it, blocking streaming on some endpoints.

    for (const endpoint of STREAMING_ENDPOINTS) {
        if (request.path.startsWith(endpoint)) {
            return false
        }
    }

    // fallback to standard filter function
    return compression.filter(request, response)
}
