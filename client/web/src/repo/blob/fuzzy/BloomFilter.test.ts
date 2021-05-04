import { BloomFilter } from "./BloomFilter";
const filter = new BloomFilter(10000, 16);
filter.add(1);
filter.add(2);
filter.add(3);

function checkNotContains(v: number) {
  test(`notContains-${v}`, () => {
    expect(filter.test(v)).toBe(false);
  });
}

function checkContains(v: number) {
  test(`contains-${v}`, () => {
    expect(filter.test(v)).toBe(true);
  });
}

checkNotContains(0);
checkContains(1);
checkContains(2);
checkContains(3);
checkNotContains(4);
