# typed-css-modules [![Build Status](https://travis-ci.org/Quramy/typed-css-modules.svg?branch=master)](https://travis-ci.org/Quramy/typed-css-modules) [![npm version](https://badge.fury.io/js/typed-css-modules.svg)](http://badge.fury.io/js/typed-css-modules)

Creates TypeScript definition files from [CSS Modules](https://github.com/css-modules/css-modules) .css files.

If you have the following css, 

```css
/* styles.css */

@value primary: red;

.myClass {
  color: primary;
}
```

typed-css-modules creates the following .d.ts files from the above css:

```ts
/* styles.css.d.ts */
export const primary: string;
export const myClass: string;
```

So, you can import CSS modules' class or variable into your TypeScript sources:

```ts
/* app.ts */
import * as styles from './styles.css';
console.log(`<div class="${styles.myClass}"></div>`);
console.log(`<div style="color: ${styles.primary}"></div>`);
```

## CLI

```sh
npm install -g typed-css-modules
```

And exec `tcm <input directory>` command. 
For example, if you have .css files under `src` directory, exec the following:

```sh
tcm src
```

Then, this creates `*.css.d.ts` files under the directory which has original .css file.

```text
(your project root)
- src/
    | myStyle.css
    | myStyle.css.d.ts [created]
```

#### output directory
Use `-o` or `--outDir` option.

For example:

```sh
tcm -o dist src
```

```text
(your project root)
- src/
    | myStyle.css
- dist/
    | myStyle.css.d.ts [created]
```

#### file name pattern

By the default, this tool searches `**/*.css` files under `<input directory>`.
If you can customize glob pattern, you can use `--pattern` or `-p` option.

```sh
tcm -p src/**/*.icss
```

#### watch
With `-w` or `--watch`, this CLI watches files in the input directory.

## API

```sh
npm install typed-css-modules
```

```js
import DtsCreator from 'typed-css-modules';
let creator = new DtsCreator();
creator.create('src/style.css').then(content => {
  console.log(content.tokens);          // ['myClass']
  console.log(content.formatted);       // 'export const myClass: string;'
  content.writeFile();                  // writes this content to "src/style.css.d.ts"
});
```

### class DtsCreator
DtsCreator instance processes the input CSS and create TypeScript definition contents.

#### `new DtsCreator(option)`
You can set the following options:

* `option.rootDir`: Project root directory(default: `process.cwd()`). 
* `option.searchDir`: Directory which includes target `*.css` files(default: `'./'`).
* `option.outDir`: Output directory(default: `option.searchDir`).

#### `create(filepath, contents) => Promise(dtsContent)`
Returns `DtsContent` instance.

* `filepath`: path of target .css file.
* `contents`(optional): the CSS content of the `filepath`. If set, DtsCreator uses the contents instead of the original contents of the `filepath`.

### class DtsContent
DtsContent instance has `*.d.ts` content, final output path, and function to write file.

#### `writeFile() => Promise(dtsContent)`
Writes the DtsContent instance's content to a file.

* `dtsContent`: the DtsContent instance itself.

#### `tokens`
An array of tokens retrieved from input CSS file.
e.g. `['myClass']`

#### `contents`
An array of TypeScript definition expressions.
e.g. `['export const myClass: string;']`.

#### `formatted`
A string of TypeScript definition expression.

e.g.

```ts
export const myClass: string;
```

#### `messageList`
An array of messages. The messages contains invalid token information.
e.g. `['my-class is not valid TypeScript variable name.']`.

#### `outputFilePath`
Final output file path.

## Remarks
If your input CSS file has the followng class names, these invalid tokens are not written to output `.d.ts` file.

```css
/* TypeScript reserved word */
.while {
  color: red;
}

/* invalid TypeScript variable */
.my-class{
  color: red;
}

/* it's ok */
.myClass {
  color: red;
}
```

## Example
There is a minimum example in this repository `example` folder. Clone this repository and run `cd example; npm i; npm start`.

Or please see [https://github.com/Quramy/typescript-css-modules-demo](https://github.com/Quramy/typescript-css-modules-demo). It's a working demonstration of CSS Modules with React and TypeScript.

## License
This software is released under the MIT License, see LICENSE.txt.
