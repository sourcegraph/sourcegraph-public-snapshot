/**
 * A document filter denotes a document by different properties like the
 * [language](#TextDocument.languageId), the scheme of its resource, or a glob-pattern that is
 * applied to the [path](#TextDocument.fileName).
 *
 * @sample A language filter that applies to typescript files on disk: `{ language: 'typescript', scheme: 'file' }`
 * @sample A language filter that applies to all package.json paths: `{ language: 'json', pattern: '**package.json' }`
 */
export type DocumentFilter =
    | {
          /** A language id, such as `typescript`. */
          language: string
          /** A URI scheme, such as `file` or `untitled`. */
          scheme?: string
          /** A glob pattern, such as `*.{ts,js}`. */
          pattern?: string
      }
    | {
          /** A language id, such as `typescript`. */
          language?: string
          /** A URI scheme, such as `file` or `untitled`. */
          scheme: string
          /** A glob pattern, such as `*.{ts,js}`. */
          pattern?: string
      }
    | {
          /** A language id, such as `typescript`. */
          language?: string
          /** A URI scheme, such as `file` or `untitled`. */
          scheme?: string
          /** A glob pattern, such as `*.{ts,js}`. */
          pattern: string
      }

export namespace DocumentFilter {
    export function is(value: any): value is DocumentFilter {
        const candidate: DocumentFilter = value
        return (
            typeof candidate.language === 'string' ||
            typeof candidate.scheme === 'string' ||
            typeof candidate.pattern === 'string'
        )
    }
}

/**
 * A document selector is the combination of one or many document filters.
 *
 * @sample `let sel:DocumentSelector = [{ language: 'typescript' }, { language: 'json', pattern: '**∕tsconfig.json' }]`;
 */
export type DocumentSelector = (string | DocumentFilter)[]
