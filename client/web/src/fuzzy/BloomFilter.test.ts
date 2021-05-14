import { BloomFilter } from './BloomFilter'
const filter = new BloomFilter(10000, 16)
filter.add(1)
filter.add(2)
filter.add(3)

function checkNotContains(number: number) {
    test(`notContains-${number}`, () => {
        expect(filter.test(number)).toBe(false)
    })
}

function checkContains(number: number) {
    test(`contains-${number}`, () => {
        expect(filter.test(number)).toBe(true)
    })
}

checkNotContains(0)
checkContains(1)
checkContains(2)
checkContains(3)
checkNotContains(4)
