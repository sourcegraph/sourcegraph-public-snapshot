import { FuzzySearch, allFuzzyParts, fuzzyMatchesQuery } from "./FuzzySearch";

const all = [
  "to/the/moon.jpg",
  "business/crazy.txt",
  "fuzzy/business.txt",
  "haha/business.txt",
  "lol/business.txt",
];
const fuzzy = new FuzzySearch(all);

function checkSearch(query: string, expected: string[]) {
  test.only(`search-${query}`, () => {
    const actual = fuzzy
      .search({ value: query, maxResults: 1000 })
      .map((t) => t.text);
    expect(actual).toStrictEqual(expected);
  });
}

function checkParts(name: string, original: string, expected: string[]) {
  test(`allFuzzyParts-${name}`, () => {
    expect(allFuzzyParts(original)).toStrictEqual(expected);
  });
}
function checkFuzzyMatch(
  name: string,
  query: string,
  value: string,
  expected: string[]
) {
  test(`fuzzyMatchesQuery-${name}`, () => {
    const obtained = fuzzyMatchesQuery(query, value);
    const parts: string[] = [];
    obtained.forEach((pos) => {
      parts.push(value.substring(pos.startOffset, pos.endOffset));
    });

    expect(parts).toStrictEqual(expected);
  });
}
function checkNoFuzzyMatch(name: string, query: string, value: string) {
  test(`!fuzzyMatchesQuery-${name}`, () => {
    expect(fuzzyMatchesQuery(query, value)).toHaveLength(0);
  });
}

checkParts("basic", "haha/business.txt", ["haha", "business", "txt"]);
checkParts("snake_case", "haha_business.txt", ["haha", "business", "txt"]);
checkParts("camelCase", "hahaBusiness.txt", ["haha", "Business", "txt"]);
checkParts("CamelCase", "HahaBusiness.txt", ["Haha", "Business", "txt"]);
checkParts("kebab-case", "haha-business.txt", ["haha", "business", "txt"]);
// checkParts("kebab-case", "haha-business.txt", ["haha", "business", "txt"]);

checkFuzzyMatch("basic", "ha/busi", "haha/business.txt", ["ha", "busi"]);

checkSearch("h/bus", ["haha/business.txt"]);
// checkSearch("moon", ["to/the/moon.jpg"]);
// checkSearch("t/moon", ["to/the/moon.jpg"]);
// checkSearch("t/t/moon", ["to/the/moon.jpg"]);
// checkSearch("t.t.moon", ["to/the/moon.jpg"]);
// checkSearch("t t moon", ["to/the/moon.jpg"]);
// checkSearch("jpg", ["to/the/moon.jpg"]);
// checkSearch("t", all);

// checkSearch("t/m", ["to/the/moon.jpg"]);
// checkSearch("mo", ["to/the/moon.jpg"]);
