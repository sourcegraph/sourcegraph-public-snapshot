/**
 * Returns a hash code value for a string.
 *
 * The hash algorithm here is using Base64 encoding of a SHA256 hash
 *
 */
export async function hashCode(input: string): Promise<string> {
    // See `SubtleCrypto` API docs:
    // https://developer.mozilla.org/en-US/docs/Web/API/SubtleCrypto/digest#converting_a_digest_to_a_hex_string
    const messageUint8 = new TextEncoder().encode(input)
    const hashBuffer = await crypto.subtle.digest('SHA-256', messageUint8)
    const hashArray = [...new Uint8Array(hashBuffer)]

    return btoa(String.fromCharCode(...hashArray))
}
