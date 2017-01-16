TypeScript style guidelines
=====
This repository uses `tslint` to enforce many best practices in TypeScript.
Internal best practices that we agree upon that can't be enforced with `tslint` should be catalogued here.

## Rules

* **access modifiers** are required for all _private_ class methods and class members. One exception is made for constructor variable assignment, below.
```TypeScript
class Foo {
    private varName: string;
    funcName() : void {}
    // not
    _varName: string;
    _funcName(): void {}
```
* **constructor variable assignent** should occur in the constructor signature, when possible
```TypeScript
class Foo {
    constructor(public varName: string) {}
    // not
    constructor(varName: string) {
        this.varName = varName;
    }
```
* **naming conventions** are as follows:
    * variables and functions are always camelCase
    * Exception: React components are PascalCase, as are functions that produce classes or React components
    * class definitions are always PascalCase
    * global constants and class constants are always UPPERCASE
    * no _ before private members
* **error logging** is accomplished with `console.error(`. Error logs are captured with Sentry, so make sure to walk that line between not logging non-errors and logging all unexpected states that we should be fixing.
* **error messages** should be lowercased, be assembled using \` notation, and terminate with a punctuation mark. Error messages should be complete sentences.
```TypeScript
// good examples
console.error(`variation ${varName} is not defined in optimizely variation list. variation list: ${varList.toString()}.`);
// bad examples
console.error(`variation not found.`);
```
* **docs should use JSDoc syntax**, which is defined at [usejsdoc.org](http://usejsdoc.org/about-getting-started.html). To emphasize, this means we should _not_ use `//` for documentation before functions. For now, defining `@parameter` and `@constructor` is optional, however.
* **undefined versus null** prefer null, except for the three cases where JavaScript uses undefined:
    * an optional argument to a function
    * a property that doesn't exist on an object
    * the output of JavaScript functions like `find`
If that isn't specific enough, defer to the theory - `undefined` means that something does not exist or was not found. `null` implies that a value is optional by design.

* When creating a new Promise, reject with an Error object, not a string.
```TypeScript
var promise = new Promise(function(resolve, reject) {
  if (/* everything turned out fine */) {
    resolve("Stuff worked!");
  }
  else {
    reject(Error("It broke")); // not reject("it broke");
  }
});
```

## Nifty tricks

Below are the collection of TypeScript shorthand and tricks that have come up in the course of code reviews.

* **using map, filter, and reduce** whenever possible is strongly encouraged.
```
arrayOfValues.map(varName => varName + 1); // this shorthand should be use when possible
arrayOfValues.forEach(methodDefinedElsewhereTakingOneArgument);
```
