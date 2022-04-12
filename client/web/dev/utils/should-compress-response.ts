import compression, { CompressionFilter } from 'compression'

export const shouldCompressResponse: CompressionFilter = (request, response) => {
    // Disable compression because gzip buffers the full response
    // before sending it, blocking streaming on some endpoints.
    if (request.path.startsWith('/search/stream')) {
        return false
    }

    if (request.path.startsWith('/.api/compute/stream')) {
        return false
    }

    // fallback to standard filter function
    return compression.filter(request, response)
}
