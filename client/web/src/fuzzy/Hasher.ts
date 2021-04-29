/**
 * Computes the hashcode from a streaming input of characters. Every hashcode is
 * computed in O(1) time.
 *
 * This class makes it possible to compute a hashcode for every prefix of a
 * given string of length N in O(N) time.  For example, given the string "Doc",
 * we can compute the hashcode for the string "D", "Do" and "Doc" in three
 * constant operations. If implemented naively, computing every individual
 * hashcode would be a linear operation resulting in a total runtime of O(N^2).
 */
export class Hasher {
    private h: number = 0
    constructor() {}
    public update(ch: string): Hasher {
        for (let i = 0; i < ch.length; i++) {
            this.h = (Math.imul(31, this.h) + ch.charCodeAt(i)) | 0
        }
        return this
    }
    public digest(): number {
        return this.h
    }
    public reset() {
        this.h = 0
    }
}
