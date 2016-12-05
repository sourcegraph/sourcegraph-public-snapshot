# commandpost [![Circle CI](https://circleci.com/gh/vvakame/commandpost.png?style=badge)](https://circleci.com/gh/vvakame/commandpost)

commandpost is a command-line option parser.
This library inspired by [commander](https://www.npmjs.com/package/commander).

commander is very user friendly, but not [TypeScript](https://www.npmjs.com/package/typescript) friendly.
commandpost is improve it.
Of course, commandpost can also be used from ordinary JavaScript program. :+1:

## Installation

```
$ npm install --save commandpost
```

## How to use

### Basic usage

```
$ cat cli.ts
import * as commandpost from "commandpost";

let root = commandpost
	.create<{ spice: string[]; }, { food: string; }>("dinner <food>")
	.version("1.0.0", "-v, --version")
	.description("today's dinner!")
	.option("-s, --spice <name>", "What spice do you want? default: pepper")
	.action((opts, args) => {
		console.log(`Your dinner is ${args.food} with ${opts.spice[0] || "pepper"}!`);
	});

commandpost
	.exec(root, process.argv)
	.catch(err => {
		if (err instanceof Error) {
			console.error(err.stack);
		} else {
			console.error(err);
		}
		process.exit(1);
	});

$ node cli.js --help
  Usage: dinner [options] [--] <food>

  Options:

    -s, --spice <name>  What spice do you want? default: pepper

$ node cli.js -s "soy sause" "fillet steak"
Your dinner is fillet steak with soy sause!
```

### Command

top level command is created by `commandpost.create` function.

commandpost can have a sub command.
sub command is created by `topLevelCommand.subCommand` method.
like [this](https://github.com/vvakame/commandpost/blob/master/example/usage.ts#L36).

commandpost can configure several items.
e.g. version information, app description, CLI usage and help message.
I recommend that you should setup `.version` and `.description`.
Usually, automatic generated help message satisfy you.

### Option

```
// shorthand style & formal style option with required parameter. option value is convert to string[].
cmd.option("-c, --config <configFile>", "Read setting from specified config file path");

// option with optional parameter. option value is convert to string[].
cmd.option("-c, --config [configFile]", "Read setting from specified config file path");

// option without parameter. option value is convert to boolean. default false.
cmd.option("--suppress-warning", "Suppress warning");

// option with `--no-` prefix. option value is convert to boolean. default true.
cmd.option("--no-type-checking", "Type checking disabled");
```

If you want to handling unknown options, You can use `.allowUnknownOption` method.

### Argument

```
// required argument
commandpost.create<{}, { food: string; }>("dinner <food>");

// optonal argument
commandpost.create<{}, { food: string; }>("dinner [food]");

// variadic argument
commandpost.create<{}, { foods: string[]; }>("dinner <food...>");
```

### Examples

* [example](https://github.com/vvakame/commandpost/blob/master/example/usage.ts) dir
* [typescript-formatter](https://github.com/vvakame/typescript-formatter/blob/master/lib/cli.ts)
* [dtsm](https://github.com/vvakame/dtsm/blob/master/lib/cli.ts)
* [review.js](https://github.com/vvakame/review.js/blob/master/lib/cli.ts)
* [prh](https://github.com/vvakame/prh/blob/master/lib/cli.ts)

## Contributing

This package's author vvakame is not native english speaker. My first language is Japanese.
If you find incorrect english, please send me a pull request.
