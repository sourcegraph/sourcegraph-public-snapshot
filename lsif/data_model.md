# LSIF data model

This document outlines the data model for a single LSIF dump. The definition of the database tables and the entities encoded within it can be found in `models.database.ts`.

In the following document, we collapse ranges to keep the document readable, where `a:b-c:d` is shorthand for the following:

```
{
    "startLine": a,
    "startCharacter": b,
    "endLine": c,
    "endCharacter": d
}
```

This applies to JSON payloads, and a similar shorthand is used for the columns of the `definitions` and `references` tables.

## Source

The following source files compose the package `sample`, which is used as the running example for this document.

**foo.ts**

```typescript
export function foo(value: string): string {
  return value.substring(1, value.length - 1)
}
```

**bar.ts**

```typescript
import { foo } from './foo'

export function bar(input: string): string {
  return foo(foo(input))
}
```

## Database values

**`meta` table**

This table is populated with **exactly** one row containing the version of the LSIF input, the version of the software that converted it into a SQLite database, and the approximate number of result chunk rows for this dump.

The last value is used in order to achieve a consistent hash of identifiers that map to the correct result chunk row identifier. This will be explained in more detail later in this document.

| id  | lsifVersion | sourcegraphVersion | numResultChunks |
| --- | ----------- | ------------------ | --------------- |
| 0   | 0.4.3       | 0.1.0              | 1               |

**`documents` table**

This table is populated with a gzipped, base64-encoded JSON payload that represents the ranges as well as each range's definition, reference, and hover result identifiers. The table is keyed on the path of the document relative to the project root.

| path   | data                         |
| ------ | ---------------------------- |
| foo.ts | _gzipped_ and _json-encoded_ |
| bar.ts | _gzipped_ and _json-encoded_ |

Each payload has the following form. As the documents are large, we show only the decoded version for `foo.ts`.

**encoded `foo.ts` payload**

````json
{
  "ranges": {
    "9": 0,
    "14": 1,
    "21": 2,
    "25": 3,
    "36": 4,
    "38": 5,
    "47": 6
  },
  "orderedRanges": [
    {
      "range": "0:0-0:0",
      "definitionResultId": "49",
      "referenceResultId": "52",
      "monikerIds": ["9007199254740990"]
    },
    {
      "range": "0:16-0:19",
      "definitionResultId": "55",
      "referenceResultId": "58",
      "hoverResultId": "16",
      "monikerIds": ["9007199254740987"]
    },
    {
      "range": "0:20-0:25",
      "definitionResultId": "61",
      "referenceResultId": "64",
      "hoverResultId": "23",
      "monikerIds": []
    },
    {
      "range": "1:9-1:14",
      "definitionResultId": "61",
      "referenceResultId": "64",
      "hoverResultId": "23",
      "monikerIds": []
    },
    {
      "range": "1:15-1:24",
      "definitionResultId": "144",
      "referenceResultId": "68",
      "hoverResultId": "34",
      "monikerIds": ["30"]
    },
    {
      "range": "1:28-1:33",
      "definitionResultId": "61",
      "referenceResultId": "64",
      "hoverResultId": "23",
      "monikerIds": []
    },
    {
      "range": "1:34-1:40",
      "definitionResultId": "148",
      "referenceResultId": "71",
      "hoverResultId": "45",
      "monikerIds": []
    }
  ],
  "hoverResults": {
    "16": "```typescript\nfunction foo(value: string): string\n```",
    "23": "```typescript\n(parameter) value: string\n```",
    "34": "```typescript\n(method) String.substring(start: number, end?: number): string\n```\n\n---\n\nReturns the substring at the specified location within a String object.",
    "45": "```typescript\n(property) String.length: number\n```\n\n---\n\nReturns the length of a String object."
  },
  "monikers": {
    "9007199254740987": {
      "kind": "export",
      "scheme": "npm",
      "identifier": "sample:foo:foo",
      "packageInformationId": "9007199254740991"
    },
    "9007199254740990": {
      "kind": "export",
      "scheme": "npm",
      "identifier": "sample:foo:",
      "packageInformationId": "9007199254740991"
    }
  },
  "packageInformation": {
    "9007199254740991": {
      "name": "sample",
      "version": "0.1.0"
    }
  }
}
````

The `ranges` field holds a map from range identifier to the index of the range data in the `orderedRanges` field. The `orderedRanges` hold the full range data and the array is sorted by the starting line and starting character. Having both fields allows us to efficiently lookup a range by identifier or by a position in the document.

Each range has optional fields for a definition result, a reference result, and a hover result. Each range also has a possibly empty list of moniker ids.

To retrieve a definition or reference result by its identifier, we must first determine in which result chunk it is defined. This requires that we take the hash of the identifier (modulo the `numResultChunks` field of the `meta` table). This gives us the unique identifier into the `resultChunks` table. The format of this payload is discussed below.

In this example, there is only one result chunk. Larger dumps will have a greater number of result chunks to keep the amount of data encoded in a single database row reasonable.

The hover result and moniker identifiers index into the `hoverResults` and `monikers` field of the document.

**`resultChunks` table**

This table is populated with gzipped, base64-encoded JSON payloads that contains a mapping from definition result or reference result identifiers to the set of ranges that compose that result. A definition or reference result may be referred to by many documents, which is why it is encoded separately. The table is keyed on the common hash of each definition and reference result id inserted in this chunk.

| id  | data                         |
| --- | ---------------------------- |
| 0   | _gzipped_ and _json-encoded_ |

Each payload has the following form.

**encoded result chunk #0 payload**

```json
{
  "documentPaths": {
    "4": "foo.ts",
    "80": "bar.ts"
  },
  "documentIdRangeIds": {
    "49": [{ "documentId": "4", "rangeId": "9" }],
    "55": [{ "documentId": "4", "rangeId": "4" }],
    "61": [{ "documentId": "4", "rangeId": "21" }],
    "71": [{ "documentId": "4", "rangeId": "47" }],
    "52": [
      { "documentId": "4", "rangeId": "9" },
      { "documentId": "80", "rangeId": "95" }
    ],
    "58": [
      { "documentId": "4", "rangeId": "14" },
      { "documentId": "80", "rangeId": "91" },
      { "documentId": "80", "rangeId": "111" },
      { "documentId": "80", "rangeId": "113" }
    ],
    "64": [
      { "documentId": "4", "rangeId": "21" },
      { "documentId": "4", "rangeId": "25" },
      { "documentId": "4", "rangeId": "38" }
    ],
    "68": [{ "documentId": "4", "rangeId": "36" }],
    "117": [{ "documentId": "80", "rangeId": "85" }],
    "120": [{ "documentId": "80", "rangeId": "85" }],
    "125": [{ "documentId": "80", "rangeId": "100" }],
    "128": [{ "documentId": "80", "rangeId": "100" }],
    "131": [{ "documentId": "80", "rangeId": "107" }],
    "134": [
      { "documentId": "80", "rangeId": "107" },
      { "documentId": "80", "rangeId": "115" }
    ]
  }
}
```

The `documentIdRangeIds` field store a list of _pairs_ of document identifiers and range identifiers. To look up a range in this format, the `documentId` must be translated into a document path via the `documentPaths` field. This gives the primary key of the document containing the range in the `documents` table, and the range identifier can be looked up in the decoded payload.

**definitions table**

This table is populated with the monikers attached to a definition result. The table is keyed on the `(scheme, identifier)` pair to allow quick lookup by moniker.

| id  | scheme | identifier     | documentPath | range        |
| --- | ------ | -------------- | ------------ | ------------ |
| 1   | npm    | sample:foo:    | foo.ts       | 0:0 to 0:0   |
| 2   | npm    | sample:foo:foo | foo.ts       | 0:16 to 0:19 |
| 3   | npm    | sample:bar:    | bar.ts       | 0:0 to 0:0   |
| 4   | npm    | sample:bar:bar | bar.ts       | 2:16 to 2:19 |

The row with id `2` correlates the `npm` moniker for the `foo` function with the range where it is defined in `foo.ts`. Similarly, the row with id `4` correlates the exported `npm` moniker for the `bar` function with the range where it is defined in `bar.ts`.

**references table**

This table is populated with the monikers attached to a reference result. The table is keyed on the `(scheme, identifier)` pair to allow quick lookup by moniker.

| id  | scheme | identifier     | documentPath | range        |
| --- | ------ | -------------- | ------------ | ------------ |
| 1   | npm    | sample:foo     | foo.ts       | 0:0 to 0:0   |
| 2   | npm    | sample:foo     | bar.ts       | 0:20 to 0:27 |
| 3   | npm    | sample:bar     | bar.ts       | 0:0 to 0:0   |
| 4   | npm    | sample:foo:foo | foo.ts       | 0:16 to 0:19 |
| 5   | npm    | sample:foo:foo | bar.ts       | 0:9 to 0:12  |
| 6   | npm    | sample:foo:foo | bar.ts       | 3:9 to 3:12  |
| 7   | npm    | sample:foo:foo | bar.ts       | 3:13 to 3:16 |
| 8   | npm    | sample:bar:bar | bar.ts       | 2:16 to 2:19 |

THe row with ids `4` through `7` correlate the `npm` moniker for the `foo` function with its references: the definition in `foo.ts`, its import in `bar.ts`, and its two uses in `bar.ts`, respectively.
