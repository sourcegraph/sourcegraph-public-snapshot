import { createHash } from "crypto";

/**
 * Returns a hash code value for a string.
 *
 * The hash algorithm here is using Base64 encoding of a SHA256 hash
 * 
 */

export function hashCode(input: string): string {

    const hash = createHash('sha256').update(input).digest('base64');  

    return hash;  
}