/**
 * Returns a hash code value for a string.
 *
 * The hash algorithm here is using Base64 encoding of a SHA256 hash
 *
 */
export async function hashCode(input: string): Promise<string> {
    // See `SubtleCrypto` API docs:
    // https://developer.mozilla.org/en-US/docs/Web/API/SubtleCrypto/digest#converting_a_digest_to_a_hex_string
    const messageUint8 = new TextEncoder().encode(input) // encode as (utf-8) Uint8Array
    const hashBuffer = await crypto.subtle.digest('SHA-256', messageUint8) // hash the message
    const hashArray = [...new Uint8Array(hashBuffer)] // convert buffer to byte array
    const base64Hash = btoa(String.fromCharCode(...hashArray))

    return base64Hash
}
