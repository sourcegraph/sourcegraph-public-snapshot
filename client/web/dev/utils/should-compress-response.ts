import compression, { CompressionFilter } from 'compression'

export const shouldCompressResponse: CompressionFilter = (request, response) => {
    // Disable compression because gzip buffers the full response
    // before sending it, which makes streaming search not stream.
    if (request.path.startsWith('/search/stream')) {
        return false
    }

    // fallback to standard filter function
    return compression.filter(request, response)
}
