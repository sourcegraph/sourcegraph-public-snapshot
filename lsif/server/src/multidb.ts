import * as lsp from 'vscode-languageserver';
import { Database, UriTransformer } from './database';
import { DocumentInfo } from './files';
import { Id } from 'lsif-protocol';
import { URI } from 'vscode-uri';

export class NamedDatabase {
  constructor(
    public name: string,
    public enabled: boolean,
    public ext: string,
    public db: Database
  ) { }
}

export class MultiDatabase extends Database {
  constructor(private databases: Array<NamedDatabase>) {
    super()
  }

  public load(file: string, transformerFactory: (projectRoot: string) => UriTransformer): Promise<void> {
    return new Promise<void>((resolve, reject) => {
      const promises = Array<Promise<NamedDatabase | null>>();
      for (const db of this.databases) {
        if (!db.enabled) {
          continue
        }

        promises.push(instrumentPromise(
          `${db.name}: load`,
          db.db.load(file + db.ext, transformerFactory).then(_ => db).catch(e => {
            if ('code' in e && e.code === 'ENOENT') {
              return null
            }

            throw e
          })
        ))
      }

      Promise.all(promises).then(dbs => {
        // Remove all the databases that failed to load due to a file in the
        // storage root either being purge or not existing in the first place
        // due to a feature flag being disabled.
        this.databases = <Array<NamedDatabase>>dbs.filter(db => !!db)

        // If we didn't load anything, raise an ENOENT to signal the server
        // and cache that we don't have any data to use for this guy.
        if (this.databases.length === 0) {
          reject({ code: 'ENOENT' })
          return
        }

        resolve()
      })
    })
  }

  public close(): void {
    this.databases.map(d => d.db.close())
  }

  // These methods are teh ones called in the http-server path. These should
  // be instrumented so that we know what performs better than the others.

  public hover(uri: string, position: lsp.Position): lsp.Hover | undefined {
    return compareResults(
      'hover',
      this.databases.map(db => instrument(`${db.name}: hover`, () => db.db.hover(uri, position)))
    )
  }

  public definitions(uri: string, position: lsp.Position): lsp.Location | lsp.Location[] | undefined {
    return compareResults(
      'definitions',
      this.databases.map(db => instrument(`${db.name}: definitions`, () => db.db.definitions(uri, position)))
    )
  }

  public references(uri: string, position: lsp.Position, context: lsp.ReferenceContext): lsp.Location[] | undefined {
    return compareResults(
      'references',
      this.databases.map(db => instrument(`${db.name}: references`, () => db.db.references(uri, position, context)))
    )
  }

  // These methods need to be defined, but are not actually called anywhere
  // in the http-server path. We don't need to do as through of a job on these.

  public getProjectRoot(): URI {
    return this.databases.map(db => db.db.getProjectRoot())[0]
  }

  public foldingRanges(uri: string): lsp.FoldingRange[] | undefined {
    return this.databases.map(db => db.db.foldingRanges(uri))[0]
  }

  public documentSymbols(uri: string): lsp.DocumentSymbol[] | undefined {
    return this.databases.map(db => db.db.documentSymbols(uri))[0]
  }

  public declarations(uri: string, position: lsp.Position): lsp.Location | lsp.Location[] | undefined {
    return this.databases.map(db => db.db.declarations(uri, position))[0]
  }

  // The following methods are abstract and need to be defined
  // but are also protected so we can just return some junk to
  // appease the compiler.

  protected getDocumentInfos(): DocumentInfo[] {
    return []
  }

  protected findFile(uri: string): Id | undefined {
    return undefined
  }

  protected fileContent(id: Id): string | undefined {
    return undefined
  }
}

//
// Helpers

function instrumentPromise<T>(name: string, p: Promise<T>): Promise<T> {
  const start = new Date().getTime()
  return p.then(res => {
    const elapsed = new Date().getTime() - start;
    console.log('%s completed in %.2fms', name, elapsed)
    return res;
  });
}

function instrument<T>(name: string, f: () => T): T {
  const start = new Date().getTime()
  const res = f();
  const elapsed = new Date().getTime() - start;
  console.log('%s completed in %.2fms', name, elapsed)
  return res;
}

function compareResults<T>(name: string, results: Array<T>): T {
  for (let i = 1; i < results.length; i++) {
    if (results[i] !== results[0]) {
      throw `Unexpected differing result from ${name}: ${results[0]} and ${results[i]}`
    }
  }

  return results[0]
}
