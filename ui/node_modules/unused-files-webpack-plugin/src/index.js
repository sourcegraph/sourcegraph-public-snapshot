import {
  join as joinPath,
} from "path";

import {
  default as glob,
} from "glob";

export class UnusedFilesWebpackPlugin {
  constructor(options = {}) {
    this.options = {
      pattern: `**/*.*`,
      ...options,
      failOnUnused: options.failOnUnused === true,
    };

    this.globOptions = {
      ignore: `node_modules/**/*`,
      ...options.globOptions,
    };
  }

  apply(compiler) {
    compiler.plugin(`after-emit`, (compilation, done) =>
      this._applyAfterEmit(compiler, compilation, done)
    );
  }

  _applyAfterEmit(compiler, compilation, done) {
    const globOptions = this._getGlobOptions(compiler);
    const fileDepsMap = this._getFileDepsMap(compilation);
    const absolutePathResolver = it => joinPath(globOptions.cwd, it);

    const handleError = err => {
      if (compilation.bail) {
        done(err);
      } else {
        compilation.errors.push(err);
      }
    };

    glob(this.options.pattern, globOptions, (err, files) => {
      if (err) {
        handleError(err);
        return;
      }
      const unused = files.filter(filepath =>
        !(absolutePathResolver(filepath) in fileDepsMap)
      );
      if (unused.length === 0) {
        done();
        return;
      }
      const error = new Error(`
UnusedFilesWebpackPlugin found some unused files:
${unused.join(`\n`)}`);

      if (this.options.failOnUnused) {
        handleError(error);
      } else {
        compilation.warnings.push(error);
        done();
      }
    });
  }

  _getGlobOptions(compiler) {
    return {
      cwd: compiler.context,
      ...this.globOptions,
    };
  }

  _getFileDepsMap(compilation) {
    const fileDepsBy = compilation.fileDependencies.reduce((acc, usedFilepath) => ({
      ...acc,
      [usedFilepath]: usedFilepath,
    }), {});

    const { assets } = compilation;
    Object.keys(assets).forEach(assetRelpath => {
      const existsAt = assets[assetRelpath].existsAt;
      fileDepsBy[existsAt] = existsAt;
    });
    return fileDepsBy;
  }
}

export default UnusedFilesWebpackPlugin;
