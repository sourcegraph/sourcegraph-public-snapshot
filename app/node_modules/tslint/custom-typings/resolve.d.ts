declare module "resolve" {
  interface ResolveOptions {
    basedir?: string;
    extensions?: string[];
    paths?: string[];
    moduleDirectory?: string | string[];
  }
  interface AsyncResolveOptions extends ResolveOptions {
    package?: any;
    readFile?: Function;
    isFile?: (file: string, cb: Function) => void;
    packageFilter?: Function;
    pathFilter?: Function;
  }
  interface SyncResolveOptions extends ResolveOptions {
    readFile?: Function;
    isFile?: (file: string) => boolean;
    packageFilter?: Function;
  }
  interface ResolveFunction {
    (id: string, cb: (err: any, res: string, pkg: any) => void): void;
    (id: string, opts: AsyncResolveOptions, cb: (err: any, res: string, pkg: any) => void): void;
    sync(id: string, opts?: SyncResolveOptions): string;
    isCore(pkg: any): any;
  }

  const resolve: ResolveFunction;
  export = resolve;
}
