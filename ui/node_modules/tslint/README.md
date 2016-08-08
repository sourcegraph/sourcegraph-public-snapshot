[![NPM version](https://badge.fury.io/js/tslint.svg)](http://badge.fury.io/js/tslint)
[![Downloads](http://img.shields.io/npm/dm/tslint.svg)](https://npmjs.org/package/tslint)
[![Circle CI](https://circleci.com/gh/palantir/tslint.svg?style=svg)](https://circleci.com/gh/palantir/tslint)
[![Join the chat at https://gitter.im/palantir/tslint](https://badges.gitter.im/Join%20Chat.svg)](https://gitter.im/palantir/tslint?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)

TSLint
======

An extensible linter for the TypeScript language.

Supports:

- custom rules
- custom formatters
- inline disabling / enabling of rules
- integration with [msbuild](https://github.com/joshuakgoldberg/tslint.msbuild), [grunt](https://github.com/palantir/grunt-tslint), [gulp](https://github.com/panuhorsmalahti/gulp-tslint), [atom](https://github.com/AtomLinter/linter-tslint), [eclipse](https://github.com/palantir/eclipse-tslint), [emacs](http://flycheck.org), [sublime](https://packagecontrol.io/packages/SublimeLinter-contrib-tslint), [vim](https://github.com/scrooloose/syntastic), [visual studio](https://visualstudiogallery.msdn.microsoft.com/6edc26d4-47d8-4987-82ee-7c820d79be1d), [vscode](https://marketplace.visualstudio.com/items?itemName=eg2.tslint), [webstorm](https://www.jetbrains.com/webstorm/help/tslint.html), and more

Table of Contents
------------

- [Installation](#installation)
- [Usage](#usage)
- [Core Rules](#core-rules)
- [Rule Flags](#rule-flags)
- [Custom Rules](#custom-rules)
- [Development](#development)
- [Creating a new release](#creating-a-new-release)


Installation
------------
<sup>[back to ToC &uarr;](#table-of-contents)</sup>

#### CLI

```
npm install -g tslint typescript
```

#### Library

```
npm install tslint typescript
```

#### Peer dependencies

`typescript` is a peer dependency of `tslint`. This allows you to update the compiler independently from the
linter. This also means that `tslint` will have to use the same version of `tsc` used to actually compile your sources.

Breaking changes in the latest dev release of `typescript@next` might break something in the linter if we haven't built against that release yet. If this happens to you, you can try:

1. picking up `tslint@next`, which may have some bugfixes not released in `tslint@latest`
   (see [release notes here](https://github.com/palantir/tslint/releases)).
2. rolling back `typescript` to a known working version.

Usage
-----
<sup>[back to ToC &uarr;](#table-of-contents)</sup>

Please ensure that the TypeScript source files compile correctly _before_ running the linter.

#### Configuration

TSLint is configured via a file named `tslint.json`. This file is loaded from the current path, or the user's home directory, in that order.

The configuration file specifies which rules are enabled and their options. These configurations may _extend_ other ones via the `"extends"` field in `tslint.json`.

```js
{
  /*
   * Possible values:
   * - the name of a built-in config
   * - the name of an NPM module which has a "main" file that exports a config object
   * - a relative path to a JSON file
   */
  "extends": "tslint:latest",
  "rules": {
    /*
     * Any rules specified here will override those from the base config we are extending
     */
    "no-constructor-vars": true
  },
  "rulesDirectory": [
    /*
     * A list of relative or absolute paths to directories that contain custom rules.
     * See the Custom Rules documentation below for more details.
     */
  ]
}
```

Built-in configs include `tslint:latest` and `tslint:recommended`. You may inspect their source [here](https://github.com/palantir/tslint/tree/master/src/configs).

__`tslint:recommended`__ is a stable, somewhat opinionated set of rules which we encourage for general TypeScript programming. This configuration follows semver, so it will _not_ have breaking changes across minor or patch releases.

__`tslint:latest`__ extends `tslint:recommended` and is continuously updated to include configuration for the latest rules in every TSLint release. Using this config may introduce breaking changes across minor releases as new rules are enabled which cause lint failures in your code. When TSLint reaches a major version bump, `tslint:recommended` will be updated to be identical to `tslint:latest`.

See the [core rules list](#core-rules) below for descriptions of all the rules.

#### CLI

usage: `tslint [options] file ...`

Options:

```
-c, --config          configuration file
--force               return status code 0 even if there are lint errors
-h, --help            display detailed help
-i, --init            generate a tslint.json config file in the current working directory
-o, --out             output file
-r, --rules-dir       rules directory
-s, --formatters-dir  formatters directory
-e, --exclude         exclude globs from path expansion
-t, --format          output format (prose, json, verbose, pmd, msbuild, checkstyle)  [default: "prose"]
--test                test that tslint produces the correct output for the specified directory
--project             path to tsconfig.json file
--type-check          enable type checking when linting a project
-v, --version         current version
```

tslint accepts the following command-line options:

```
-c, --config:
    The location of the configuration file that tslint will use to
    determine which rules are activated and what options to provide
    to the rules. If no option is specified, the config file named
    tslint.json is used, so long as it exists in the path.
    The format of the file is { rules: { /* rules list */ } },
    where /* rules list */ is a key: value comma-separated list of
    rulename: rule-options pairs. Rule-options can be either a
    boolean true/false value denoting whether the rule is used or not,
    or a list [boolean, ...] where the boolean provides the same role
    as in the non-list case, and the rest of the list are options passed
    to the rule that will determine what it checks for (such as number
    of characters for the max-line-length rule, or what functions to ban
    for the ban rule).

-e, --exclude:
    A filename or glob which indicates files to exclude from linting.
    This option can be supplied multiple times if you need multiple
    globs to indicate which files to exclude.

--force:
    Return status code 0 even if there are any lint errors.
    Useful while running as npm script.

-i, --init:
    Generates a tslint.json config file in the current working directory.

-o, --out:
    A filename to output the results to. By default, tslint outputs to
    stdout, which is usually the console where you're running it from.

-r, --rules-dir:
    An additional rules directory, for user-created rules.
    tslint will always check its default rules directory, in
    node_modules/tslint/lib/rules, before checking the user-provided
    rules directory, so rules in the user-provided rules directory
    with the same name as the base rules will not be loaded.

-s, --formatters-dir:
    An additional formatters directory, for user-created formatters.
    Formatters are files that will format the tslint output, before
    writing it to stdout or the file passed in --out. The default
    directory, node_modules/tslint/build/formatters, will always be
    checked first, so user-created formatters with the same names
    as the base formatters will not be loaded.

-t, --format:
    The formatter to use to format the results of the linter before
    outputting it to stdout or the file passed in --out. The core
    formatters are prose (human readable), json (machine readable)
    and verbose. prose is the default if this option is not used.
    Other built-in options include pmd, msbuild, checkstyle, and vso.
    Additional formatters can be added and used if the --formatters-dir
    option is set.

--test:
    Runs tslint on the specified directory and checks if tslint's output matches
    the expected output in .lint files. Automatically loads the tslint.json file in the
    specified directory as the configuration file for the tests. See the
    full tslint documentation for more details on how this can be used to test custom rules.

--project:
    The location of a tsconfig.json file that will be used to determine which
    files will be linted.

--type-check
    Enables the type checker when running linting rules. --project must be
    specified in order to enable type checking.

-v, --version:
    The current version of tslint.

-h, --help:
    Prints this help message.
```

#### Library

```javascript
const Linter = require("tslint");
const fs = require("fs");

const fileName = "Specify file name";
const configuration = {
    rules: {
        "variable-name": true,
        "quotemark": [true, "double"]
    }
};
const options = {
    formatter: "json",
    configuration: configuration,
    rulesDirectory: "customRules/",
    formattersDirectory: "customFormatters/"
};

const fileContents = fs.readFileSync(fileName, "utf8");
const linter = new Linter(fileName, fileContents, options);
const result = linter.lint();
```

#### Type Checking

To enable rules that work with the type checker, a TypeScript program object must be passed to the linter when using the programmatic API. Helper functions are provided to create a program from a `tsconfig.json` file. A project directory can be specified if project files do not lie in the same directory as the `tsconfig.json` file.

```javascript
const program = Linter.createProgram("tsconfig.json", "projectDir/");
const files = Linter.getFileNames(program);
const results = files.map(file => {
    const fileContents = program.getSourceFile(file).getFullText();
    const linter = new Linter(file, fileContents, options, program);
    return result.lint();
});
```

When using the CLI, the `--project` flag will automatically create a program from the specified `tsconfig.json` file. Adding `--type-check` then enables rules that require the type checker.


Core Rules
-----
<sup>[back to ToC &uarr;](#table-of-contents)</sup>

Core rules are included in the `tslint` package.

* `align` enforces vertical alignment. Rule options:
  * `"parameters"` checks alignment of function parameters.
  * `"arguments"` checks alignment of function call arguments.
  * `"statements"` checks alignment of statements.
* `arrow-parens` requires parentheses around the parameters of arrow function definitions.
* `ban` bans the use of specific functions. Options are `["object", "function"]` pairs that ban the use of `object.function()`.  An optional 3rd parameter may be provided (`["object", "function", "Use 'object.otherFunc' instead."]`) to offer an explanation as to why the function has been banned or to offer an alternative.
* `class-name` enforces PascalCased class and interface names.
* `comment-format` enforces rules for single-line comments. Rule options:
    * `"check-space"` enforces the rule that all single-line comments must begin with a space, as in `// comment`
        * note that comments starting with `///` are also allowed, for things such as `///<reference>`
    * `"check-lowercase"` enforces the rule that the first non-whitespace character of a comment must be lowercase, if applicable.
    * `"check-uppercase"` enforces the rule that the first non-whitespace character of a comment must be uppercase, if applicable.
* `curly` enforces braces for `if`/`for`/`do`/`while` statements.
* `eofline` enforces the file to end with a newline.
* `forin` enforces a `for ... in` statement to be filtered with an `if` statement.
* `indent` enforces indentation with tabs or spaces. Rule options (one is required):
    * `"tabs"` enforces consistent tabs.
    * `"spaces"` enforces consistent spaces.
* `interface-name` enforces consistent interface names. Rule options:
    * `"always-prefix"` enforces interface names must have an 'I' prefix
    * `"never-prefix"` enforces interface names must not have an 'I' prefix
* `jsdoc-format` enforces basic format rules for jsdoc comments -- comments starting with `/**`
    * each line contains an asterisk and asterisks must be aligned
    * each asterisk must be followed by either a space or a newline (except for the first and the last)
    * the only characters before the asterisk on each line must be whitespace characters
    * one line comments must start with `/** ` and end with ` */`
* `label-position` enforces labels only on sensible statements.
* `label-undefined` checks that labels are defined before usage.
* `linebreak-style` checks that line breaks used in source files are either linefeed or carriage-return linefeeds. By default linefeeds are required. This rule accepts one parameter, either "LF" or "CRLF".
* `max-file-line-count` sets the maximum number of lines for files.
* `max-line-length` sets the maximum length of a line.
* `member-access` enforces using explicit visibility on class members
    * `"check-accessor"` enforces explicit visibility on get/set accessors
    * `"check-constructor"` enforces explicit visibility on constructors
* `member-ordering` enforces member ordering. The first option should be an object with an `order` key.
   Values for `order` can be `fields-first`, `statics-first`, `instance-sandwich`, or a custom order.
* `new-parens` enforces parentheses when invoking a constructor via the `new` keyword.
* `no-angle-bracket-type-assertion` disallows usages of `<>` type assertions in favor of using the `as` keyword.
* `no-any` diallows usages of `any` as a type decoration.
* `no-arg` disallows access to `arguments.callee`.
* `no-bitwise` disallows bitwise operators.
* `no-conditional-assignment` disallows any type of assignment in any conditionals. This applies to `do-while`, `for`, `if`, and `while` statements.
* `no-consecutive-blank-lines` disallows having more than one blank line in a row in a file.
* `no-console` disallows access to the specified functions on `console`. Rule options are functions to ban on the console variable.
* `no-construct` disallows access to the constructors of `String`, `Number`, and `Boolean`.
* `no-constructor-vars` disallows the `public` and `private` modifiers for constructor parameters.
* `no-debugger` disallows `debugger` statements.
* `no-default-export` disallows default exports in ES6-style modules. Use named exports instead.
* `no-duplicate-key` disallows duplicate keys in object literals.
* `no-duplicate-variable` disallows duplicate variable declarations in the same block scope.
* `no-empty` disallows empty blocks.
* `no-eval` disallows `eval` function invocations.
* `no-for-in-array` disallows iterating over an array with a for-in loop (requires type checking).
* `no-inferrable-types` disallows explicit type declarations for variables or parameters initialized to a number, string, or boolean.
   * `ignore-params` allows specifying an inferrable type as a function param
* `no-internal-module` disallows internal `module` (use `namespace` instead).
* `no-invalid-this` disallows using the `this` keyword outside of classes.
    * `check-function-in-method` disallows using the `this` keyword in functions within class methods.
* `no-mergeable-namespace` disallows mergeable namespaces in the same file.
* `no-namespace` disallows both internal `module`s and `namespace`, but allows ES6-style external modules.
    * `allow-declarations` allows `declare namespace ... {}` to describe external APIs.
* `no-null-keyword` disallows use of the `null` keyword literal.
* `no-reference` disallows `/// <reference path=>` imports (use ES6-style imports instead).
* `no-require-imports` disallows invocation of `require()` (use ES6-style imports instead).
* `no-shadowed-variable` disallows shadowed variable declarations.
* `no-string-literal` disallows object access via string literals.
* `no-switch-case-fall-through` disallows falling through case statements. As of TypeScript version 1.8, this rule can be enabled within the compiler by passing the `--noFallthroughCasesInSwitch` flag.
* `no-trailing-whitespace` disallows trailing whitespace at the end of a line.
* `no-unreachable` disallows unreachable code after `break`, `catch`, `throw`, and `return` statements. This rule is supported and enforced by default within the TypeScript compiler since version 1.8.
* `no-unused-expression` disallows unused expression statements, that is, expression statements that are not assignments or function invocations (and thus no-ops). Combine with `no-unused-new` to disallow expressions containing the new keyword.
* `no-unused-new` disallows unused expressions statements which include the new keyword.
* `no-unused-variable` disallows unused imports, variables, functions and private class members. Rule options:
    * `"check-parameters"` disallows unused function and constructor parameters.
        * NOTE: this option is experimental and does not work with classes that use abstract method declarations, among other things. Use at your own risk.
    * `"react"` relaxes the rule for a namespace import named `React` (from either the module `"react"` or `"react/addons"`). Any JSX expression in the file will be treated as a usage of `React` (because it expands to `React.createElement`).
    * `{"ignore-pattern": "pattern"}` where pattern is a case-sensitive regexp. Variable names that match the pattern will be ignored.
* `no-use-before-declare` disallows usage of variables before their declaration.
* `no-var-keyword` disallows usage of the `var` keyword, use `let` or `const` instead.
* `no-var-requires` disallows the use of require statements except in import statements, banning the use of forms such as `var module = require("module")`.
* `object-literal-sort-keys` checks that keys in object literals are declared in alphabetical order (useful to prevent merge conflicts).
* `one-line` enforces the specified tokens to be on the same line as the expression preceding it. Rule options:
  * `"check-catch"` checks that `catch` is on the same line as the closing brace for `try`.
  * `"check-else"` checks that `else` is on the same line as the closing brace for `if`.
  * `"check-finally"` checks that `finally` is on the same line as the closing brace for the preceding `try` or `catch`.
  * `"check-open-brace"` checks that an open brace falls on the same line as its preceding expression.
  * `"check-whitespace"` checks preceding whitespace for the specified tokens.
* `one-variable-per-declaration` disallows multiple variable definitions in the same statement.
  * `"ignore-for-loop"` allows multiple variable definitions in for loop statement.
* `only-arrow-functions` disallows traditional `function () { ... }` declarations, preferring `() => { ... }` arrow lambdas.
* `quotemark` enforces consistent single or double quoted string literals. Rule options (at least one of `"double"` or `"single"` is required):
    * `"single"` enforces single quotes.
    * `"double"` enforces double quotes.
    * `"jsx-single"` enforces single quotes for JSX attributes.
    * `"jsx-double"` enforces double quotes for JSX attributes.
    * `"avoid-escape"` allows you to use the "other" quotemark in cases where escaping would normally be required. For example, `[true, "double", "avoid-escape"]` would not report a failure on the string literal `'Hello "World"'`.
* `radix` enforces the radix parameter of `parseInt`.
* `restrict-plus-operands` enforces the type of addition operands to be both `string` or both `number` (requires type checking).
* `semicolon` enforces consistent semicolon usage at the end of every statement. Rule options:
    * `"always"` enforces semicolons at the end of every statement.
    * `"never"` disallows semicolons at the end of every statement except for when they are necessary.
* `switch-default` enforces a `default` case in `switch` statements.
* `trailing-comma` enforces or disallows trailing comma within array and object literals, destructuring assignment and named imports.
  Each rule option requires a value of `"always"` or `"never"`. Rule options:
    * `"multiline"` checks multi-line object literals.
    * `"singleline"` checks single-line object literals.
* `triple-equals` enforces `===` and `!==` in favor of `==` and `!=`.
    * `"allow-null-check"` allows `==` and `!=` when comparing to `null`.
    * `"allow-undefined-check"` allows `==` and `!=` when comparing to `undefined`.
* `typedef` enforces type definitions to exist. Rule options:
    * `"call-signature"` checks return type of non-arrow functions.
    * `"arrow-call-signature"` checks return type of arrow functions.
    * `"parameter"` checks type specifier of function parameters for non-arrow functions.
    * `"arrow-parameter"` checks type specifier of function parameters for arrow functions.
    * `"property-declaration"` checks return types of interface properties.
    * `"variable-declaration"` checks variable declarations.
    * `"member-variable-declaration"` checks member variable declarations. For arrow functions being assigned as properties, either the property itself or the arrow functions parameters must have a typedef.
* `typedef-whitespace` enforces spacing whitespace for type definitions. Each rule option requires a value of `"nospace"`,
  `"onespace"` or `"space"` to require no space, exactly one or at least one space before or after the type specifier's
  colon. You can specify two objects containing the five options. The first one describes the left, the second one the
  right hand side of the typedef colon. To omit checks for either side, omit the second object or pass an empty object
  for the first. Rule options:
    * `"call-signature"` checks return type of functions.
    * `"index-signature"` checks index type specifier of indexers.
    * `"parameter"` checks function parameters.
    * `"property-declaration"` checks object property declarations.
    * `"variable-declaration"` checks variable declaration.
* `use-isnan` enforces that you use the isNaN() function to check for NaN references instead of a comparison to the NaN constant.
* `use-strict` enforces ECMAScript 5's strict mode.
    * `check-module` checks that all top-level modules are using strict mode.
    * `check-function` checks that all top-level functions are using strict mode.
* `variable-name` checks variables names for various errors.  Rule options:
  * `"check-format"`: allows only camelCased or UPPER_CASED variable names
    * `"allow-leading-underscore"` allows underscores at the beginning.
    * `"allow-trailing-underscore"` allows underscores at the end.
    * `"allow-pascal-case"` allows PascalCase in addition to camelCase.
  * `"ban-keywords"`: disallows the use of certain TypeScript keywords (`any`, `Number`, `number`, `String`, `string`, `Boolean`, `boolean`, `undefined`) as variable or parameter names.
* `whitespace` enforces spacing whitespace. Rule options:
  * `"check-branch"` checks branching statements (`if`/`else`/`for`/`while`) are followed by whitespace.
  * `"check-decl"`checks that variable declarations have whitespace around the equals token.
  * `"check-module"` checks for whitespace in import & export statements.
  * `"check-operator"` checks for whitespace around operator tokens.
  * `"check-separator"` checks for whitespace after separator tokens (`,`/`;`).
  * `"check-type"` checks for whitespace before a variable type specification.
  * `"check-typecast"` checks for whitespace between a typecast and its target.

Rule Flags
-----
<sup>[back to ToC &uarr;](#table-of-contents)</sup>

You may enable/disable TSLint or a subset of rules within certain lines of a file with the following comment rule flags:

* `/* tslint:disable */` - Disable all rules for the rest of the file
* `/* tslint:enable */` - Enable all rules for the rest of the file
* `/* tslint:disable:rule1 rule2 rule3... */` - Disable the listed rules for the rest of the file
* `/* tslint:enable:rule1 rule2 rule3... */` - Enable the listed rules for the rest of the file
* `// tslint:disable-next-line` - Disables all rules for the following line
* `someCode(); // tslint:disable-line` - Disables all rules for the current line
* `// tslint:disable-next-line:rule1 rule2 rule3...` - Disables the listed rules for the next line
* etc.

Rules flags enable or disable rules as they are parsed. Disabling an already disabled rule or enabling an already enabled rule has no effect.

For example, imagine the directive `/* tslint:disable */` on the first line of a file, `/* tslint:enable:ban class-name */` on the 10th line and `/* tslint:enable */` on the 20th. No rules will be checked between the 1st and 10th lines, only the `ban` and `class-name` rules will be checked between the 10th and 20th, and all rules will be checked for the remainder of the file.

Custom Rules
------------
<sup>[back to ToC &uarr;](#table-of-contents)</sup>

#### Custom rule sets from Palantir

- [tslint-react](https://github.com/palantir/tslint-react) - Lint rules related to React & JSX.

#### Custom rule sets from the community

If we don't have all the rules you're looking for, you can either write your own custom rules or use custom rules that others have developed. The repos below are a good source of custom rules:

- [ESLint rules for TSLint](https://github.com/buzinas/tslint-eslint-rules) - Improve your TSLint with the missing ESLint Rules
- [tslint-microsoft-contrib](https://github.com/Microsoft/tslint-microsoft-contrib) - A set of TSLint rules used on some Microsoft projects
- [codelyzer](https://github.com/mgechev/codelyzer) - A set of tslint rules for static code analysis of Angular 2 TypeScript projects
- [vrsource-tslint-rules](https://github.com/vrsource/vrsource-tslint-rules)

#### Writing custom rules

TSLint ships with a set of core rules that can be configured. However, users are also allowed to write their own rules, which allows them to enforce specific behavior not covered by the core of TSLint. TSLint's internal rules are itself written to be pluggable, so adding a new rule is as simple as creating a new rule file named by convention. New rules can be written in either TypeScript or JavaScript; if written in TypeScript, the code must be compiled to JavaScript before invoking TSLint.

Rule names are always camel-cased and *must* contain the suffix `Rule`. Let us take the example of how to write a new rule to forbid all import statements (you know, *for science*). Let us name the rule file `noImportsRule.ts`. Rules can be referenced in `tslint.json` in their kebab-case forms, so `"no-imports": true` would turn on the rule.

Now, let us first write the rule in TypeScript. A few things to note:
- We import `tslint/lib/lint` to get the whole `Lint` namespace instead of just the `Linter` class.
- The exported class must always be named `Rule` and extend from `Lint.Rules.AbstractRule`.

```typescript
import * as ts from "typescript";
import * as Lint from "tslint/lib/lint";

export class Rule extends Lint.Rules.AbstractRule {
    public static FAILURE_STRING = "import statement forbidden";

    public apply(sourceFile: ts.SourceFile): Lint.RuleFailure[] {
        return this.applyWithWalker(new NoImportsWalker(sourceFile, this.getOptions()));
    }
}

// The walker takes care of all the work.
class NoImportsWalker extends Lint.RuleWalker {
    public visitImportDeclaration(node: ts.ImportDeclaration) {
        // create a failure at the current position
        this.addFailure(this.createFailure(node.getStart(), node.getWidth(), Rule.FAILURE_STRING));

        // call the base version of this visitor to actually parse this node
        super.visitImportDeclaration(node);
    }
}
```

Given a walker, TypeScript's parser visits the AST using the visitor pattern. So the rule walkers only need to override the appropriate visitor methods to enforce its checks. For reference, the base walker can be found in [syntaxWalker.ts](https://github.com/palantir/tslint/blob/master/src/language/walker/syntaxWalker.ts).

We still need to hook up this new rule to TSLint. First make sure to compile `noImportsRule.ts`:

```bash
tsc -m commonjs --noImplicitAny noImportsRule.ts node_modules/tslint/lib/tslint.d.ts
```

Then, if using the CLI, provide the directory that contains this rule as an option to `--rules-dir`. If using TSLint as a library or via `grunt-tslint`, the `options` hash must contain `"rulesDirectory": "..."`. If you run the linter, you'll see that we have now successfully banned all import statements via TSLint!

Final notes:

- Core rules cannot be overwritten with a custom implementation.
- Custom rules can also take in options just like core rules (retrieved via `this.getOptions()`).

Custom Formatters
-----------------
<sup>[back to ToC &uarr;](#table-of-contents)</sup>

Just like rules, additional formatters can also be supplied to TSLint via `--formatters-dir` on the CLI or `formattersDirectory` option on the library or `grunt-tslint`. Writing a new formatter is simpler than writing a new rule, as shown in the JSON formatter's code.

```typescript
import * as ts from "typescript";
import * as Lint from "tslint/lib/lint";

export class Formatter extends Lint.Formatters.AbstractFormatter {
    public format(failures: Lint.RuleFailure[]): string {
        var failuresJSON = failures.map((failure: Lint.RuleFailure) => failure.toJson());
        return JSON.stringify(failuresJSON);
    }
}
```

Such custom formatters can also be written in JavaScript. Formatter files are always named with the suffix `Formatter` and the exported class within the file must be named `Formatter`. A formatter is referenced from TSLint without its suffix.

Development
-----------
<sup>[back to ToC &uarr;](#table-of-contents)</sup>

#### Quick Start

```bash
git clone git@github.com:palantir/tslint.git
npm install
grunt
```

#### `next` branch

The [`next` branch of this repo](https://github.com/palantir/tslint/tree/next) tracks the latest TypeScript compiler
nightly release as a `peerDependency`. This allows you to develop the linter and its rules against the latest features of the
language. Releases from this branch are published to npm with the `next` dist-tag, so you may install the latest dev
version of TSLint via `npm install tslint@next`.

Creating a new release
----------------------
<sup>[back to ToC &uarr;](#table-of-contents)</sup>

1. Bump the version number in `package.json` and `src/tslint.ts`
2. Add release notes in `CHANGELOG.md`
3. Run `grunt` to build the latest sources
4. Commit with message `Prepare release <version>`
5. Run `npm publish`
6. Create a git tag for the new release and push it ([see existing tags here](https://github.com/palantir/tslint/tags))
