/* eslint-disable */
/**
 * ===============================================================
 * ORIGINAL LICENSE
 * https://github.com/jasondavies/bloomfilter.js/blob/649e43e60ded806a3f0d1ca88d6a42c6dffba2db/LICENSE
 * This file has been adapted toG
 * ===============================================================
 * Copyright (c) 2018, Jason Davies
 * All rights reserved.
 *
 * Redistribution and use in source and binary forms, with or without
 * modification, are permitted provided that the following conditions are met:
 *
 * * Redistributions of source code must retain the above copyright notice, this
 *   list of conditions and the following disclaimer.
 *
 * * Redistributions in binary form must reproduce the above copyright notice,
 *   this list of conditions and the following disclaimer in the documentation
 *   and/or other materials provided with the distribution.
 *
 * * Neither the name of the copyright holder nor the names of its
 *   contributors may be used to endorse or promote products derived from
 *   this software without specific prior written permission.
 *
 * THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
 * AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
 * IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
 * DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE
 * FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL
 * DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR
 * SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER
 * CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY,
 * OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
 * OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
 */
;(function (exports) {
  exports.BloomFilter = BloomFilter
  exports.fnv_1a = fnv_1a

  // Creates a new bloom filter.  If *m* is an array-like object, with a length
  // property, then the bloom filter is loaded with data from the array, where
  // each element is a 32-bit integer.  Otherwise, *m* should specify the
  // number of bits.  Note that *m* is rounded up to the nearest multiple of
  // 32.  *k* specifies the number of hashing functions.
  function BloomFilter(m, k) {
    var a
    if (typeof m !== 'number') (a = m), (m = a.length * 32)
    var n = Math.ceil(m / 32),
      i = -1
    this.m = m = n * 32
    this.k = k

    var kbytes = 1 << Math.ceil(Math.log(Math.ceil(Math.log(m) / Math.LN2 / 8)) / Math.LN2),
      array = kbytes === 1 ? Uint8Array : kbytes === 2 ? Uint16Array : Uint32Array,
      kbuffer = new ArrayBuffer(kbytes * k)
    this.buckets = new Int32Array(n)
    if (a) while (++i < n) this.buckets[i] = a[i]
    this._locations = new array(kbuffer)
  }

  // See http://willwhim.wpengine.com/2011/09/03/producing-n-hash-functions-by-hashing-only-once/
  BloomFilter.prototype.locations = function (v) {
    var k = this.k,
      m = this.m,
      r = this._locations,
      a = fnv_1a(v),
      b = fnv_1a(v, 1576284489), // The seed value is chosen randomly
      x = a % m
    for (var i = 0; i < k; ++i) {
      r[i] = x < 0 ? x + m : x
      x = (x + b) % m
    }
    return r
  }

  BloomFilter.prototype.add = function (v) {
    var l = this.locations(v),
      k = this.k,
      buckets = this.buckets
    for (var i = 0; i < k; ++i) buckets[Math.floor(l[i] / 32)] |= 1 << l[i] % 32
  }

  BloomFilter.prototype.test = function (v) {
    var l = this.locations(v),
      k = this.k,
      buckets = this.buckets
    for (var i = 0; i < k; ++i) {
      var b = l[i]
      if ((buckets[Math.floor(b / 32)] & (1 << b % 32)) === 0) {
        return false
      }
    }
    return true
  }

  // Estimated cardinality.
  BloomFilter.prototype.size = function () {
    var buckets = this.buckets,
      bits = 0
    for (var i = 0, n = buckets.length; i < n; ++i) bits += popcnt(buckets[i])
    return (-this.m * Math.log(1 - bits / this.m)) / this.k
  }

  // http://graphics.stanford.edu/~seander/bithacks.html#CountBitsSetParallel
  function popcnt(v) {
    v -= (v >> 1) & 0x55555555
    v = (v & 0x33333333) + ((v >> 2) & 0x33333333)
    return (((v + (v >> 4)) & 0xf0f0f0f) * 0x1010101) >> 24
  }

  // Fowler/Noll/Vo hashing.
  // Nonstandard variation: this function optionally takes a seed value that is incorporated
  // into the offset basis. According to http://www.isthe.com/chongo/tech/comp/fnv/index.html
  // "almost any offset_basis will serve so long as it is non-zero".
  function fnv_1a(v, seed) {
    var a = 2166136261 ^ (seed || 0)
    var c = v,
      d = c & 0xff00
    if (d) a = fnv_multiply(a ^ (d >> 8))
    a = fnv_multiply(a ^ (c & 0xff))
    return fnv_mix(a)
  }

  // a * 16777619 mod 2**32
  function fnv_multiply(a) {
    return a + (a << 1) + (a << 4) + (a << 7) + (a << 8) + (a << 24)
  }

  // See https://web.archive.org/web/20131019013225/http://home.comcast.net/~bretm/hash/6.html
  function fnv_mix(a) {
    a += a << 13
    a ^= a >>> 7
    a += a << 3
    a ^= a >>> 17
    a += a << 5
    return a & 0xffffffff
  }
})(typeof exports !== 'undefined' ? exports : this)
