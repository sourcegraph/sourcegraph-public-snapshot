/**
 * Returns a hash code value for a string.
 *
 * The hash algorithm here is similar to the one used in Java's String.hashCode
 * and in lsifstore.
 */
export async function hashCode(string: string, maxIndex: number): Promise<string> {
    const crypto = await require('crypto');
    const secret = maxIndex.toString();
    const hash = crypto.createHmac('sha256', secret).update(string).digest('hex');  

    return hash;  
}
