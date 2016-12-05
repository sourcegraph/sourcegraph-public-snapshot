# TypeScript Formatter (tsfmt) [![Build Status](https://travis-ci.org/vvakame/typescript-formatter.svg)](https://travis-ci.org/vvakame/typescript-formatter) [![Dependency Status](https://david-dm.org/vvakame/typescript-formatter.svg?theme=shields.io)](https://david-dm.org/vvakame/typescript-formatter)

A TypeScript code formatter powered by [TypeScript Compiler Service](https://github.com/Microsoft/TypeScript/wiki/Using-the-Compiler-API#pretty-printer-using-the-ls-formatter).

```bash
$ tsfmt --help
  Usage: tsfmt [options] [--] [files...]

  Options:

    -r, --replace      replace .ts file
    --verify           checking file format
    --baseDir <path>   config file lookup from <path>
    --stdin            get formatting content from stdin
    --no-tsconfig      don't read a tsconfig.json
    --no-tslint        don't read a tslint.json
    --no-editorconfig  don't read a .editorconfig
    --no-tsfmt         don't read a tsfmt.json
    --verbose          makes output more verbose
```

## Installation

```npm install -g typescript-formatter```

## Usage

### Format or verify specific TypeScript files

```bash
$ cat sample.ts
class Sample {hello(word="world"){return "Hello, "+word;}}
new Sample().hello("TypeScript");
```

```bash
# basic. read file, output to stdout.
$ tsfmt sample.ts
class Sample { hello(word = "world") { return "Hello, " + word; } }
new Sample().hello("TypeScript");
```

```bash
# from stdin. read from stdin, output to stdout.
$ cat sample.ts | tsfmt --stdin
class Sample { hello(word = "world") { return "Hello, " + word; } }
new Sample().hello("TypeScript");
```

```bash
# replace. read file, and replace file.
$ tsfmt -r sample.ts
replaced sample.ts
$ cat sample.ts
class Sample { hello(word = "world") { return "Hello, " + word; } }
new Sample().hello("TypeScript");
```

```bash
# verify. checking file format.
$ tsfmt --verify sample.ts
sample.ts is not formatted
$ echo $?
1
```

### Reformat all files in a TypeScript project

If no files are specified on the command line but
a TypeScript project file (tsconfig.json) exists,
the list of files will be read from the project file.

```bash
# reads list of files to format from tsconfig.json
tsfmt -r
```

## Note

now `indentSize` parameter is ignored. it is TypeScript compiler matters.

## Read Settings From Files

1st. Read settings from tsfmt.json. Bellow is the example with [default values](https://github.com/vvakame/typescript-formatter/blob/master/lib/utils.ts):

```json
{
  "indentSize": 4,
  "tabSize": 4,
  "newLineCharacter": "\r\n",
  "convertTabsToSpaces": true,
  "insertSpaceAfterCommaDelimiter": true,
  "insertSpaceAfterSemicolonInForStatements": true,
  "insertSpaceBeforeAndAfterBinaryOperators": true,
  "insertSpaceAfterKeywordsInControlFlowStatements": true,
  "insertSpaceAfterFunctionKeywordForAnonymousFunctions": false,
  "insertSpaceAfterOpeningAndBeforeClosingNonemptyParenthesis": false,
  "insertSpaceAfterOpeningAndBeforeClosingNonemptyBrackets": false,
  "insertSpaceAfterOpeningAndBeforeClosingTemplateStringBraces": false,
  "placeOpenBraceOnNewLineForFunctions": false,
  "placeOpenBraceOnNewLineForControlBlocks": false
}

```

2nd. Read settings from tsconfig.json ([tsconfig.json](https://www.typescriptlang.org/docs/handbook/tsconfig-json.html))

```text
{
  "compilerOptions": {
    "newLine": "LF"
  }
}
```

3rd. Read settings from .editorconfig ([editorconfig](http://editorconfig.org/))

```text
# EditorConfig is awesome: http://EditorConfig.org

# top-most EditorConfig file
root = true

# Unix-style newlines with a newline ending every file
[*]
indent_style = tab
tab_width = 2
end_of_line = lf
charset = utf-8
trim_trailing_whitespace = true
insert_final_newline = true
```

4th. Read settings from tslint.json ([tslint](https://www.npmjs.org/package/tslint))

```json
{
  "rules": {
    "indent": [true, 4],
    "whitespace": [true,
      "check-branch",
      "check-operator",
      "check-separator"
    ]
  }
}
```

### Read Settings Rules

```
$ tree -a
.
├── foo
│   ├── bar
│   │   ├── .editorconfig
│   │   └── buzz.ts
│   ├── fuga
│   │   ├── piyo.ts
│   │   └── tsfmt.json
│   └── tsfmt.json
└── tslint.json

3 directories, 6 files
```

1. exec `$ tsfmt -r foo/bar/buzz.ts foo/fuga/piyo.ts`
2. for foo/bar/buzz.ts, read foo/tsfmt.json and foo/bar/.editorconfig and ./tslint.json
3. for foo/fuga/piyo.ts, read foo/fuga/tsfmt.json and ./tslint.json

## Change Log

See [CHANGELOG](https://github.com/vvakame/typescript-formatter/blob/master/CHANGELOG.md)
