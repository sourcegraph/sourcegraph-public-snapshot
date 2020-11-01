/**
 * Returns a hash code value for a string.
 *
 * The hash algorithm here is using Base64 encoding of a SHA256 hash
 *
 */
export async function hashCode(input: string): Promise<string> {
    const messageUint8 = new TextEncoder().encode(input)
    const hashBuffer = await crypto.subtle.digest('SHA-256', messageUint8)
    const hashArray = [...new Uint8Array(hashBuffer)]
    return hashArray.map(b => b.toString(16).padStart(2, '0')).join('')
}
