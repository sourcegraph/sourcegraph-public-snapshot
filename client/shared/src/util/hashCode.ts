import sjcl from 'sjcl'

/**
 * Returns a hash code value for a string.
 *
 * The hash algorithm here is using Base64 encoding of a SHA256 hash
 *
 */
export function hashCode(input: string): string {
    const hashBitArray = sjcl.hash.sha256.hash(input)

    return sjcl.codec.base64.fromBits(hashBitArray)
}
