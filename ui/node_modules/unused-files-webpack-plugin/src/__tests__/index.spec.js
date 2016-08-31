import {
  resolve as resolvePath,
} from "path";

import {
  default as test,
} from "tape";

import {
  default as MemoryFS,
} from "memory-fs";

import {
  default as webpack,
} from "webpack";

import {
  default as UnusedFilesWebpackPlugin,
} from "../index";

const EXPECTED_FILENAME_LIST = [
  `CHANGELOG.md`,
  `README.md`,
  `src/__tests__/index.spec.js`,
  `package.json`,
];

test(`UnusedFilesWebpackPlugin`, t => {
  const compiler = webpack({
    context: resolvePath(__dirname, `../../`),
    entry: {
      UnusedFilesWebpackPlugin: resolvePath(__dirname, `../index.js`),
    },
    output: {
      path: __dirname, // It will be in MemoryFS :)
    },
    plugins: [new UnusedFilesWebpackPlugin()],
  });
  compiler.outputFileSystem = new MemoryFS();

  compiler.run((err, stats) => {
    t.equal(err, null);

    const { warnings } = stats.compilation;
    t.equal(warnings.length, 1);

    const [unusedFilesError] = warnings;
    t.equal(unusedFilesError instanceof Error, true);

    const { message } = unusedFilesError;
    const containsExpected = EXPECTED_FILENAME_LIST.every(filename =>
      message.match(filename)
    );
    t.equal(containsExpected, true);

    t.end();
  });
});
