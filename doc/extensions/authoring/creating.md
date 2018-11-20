# Creating a Sourcegraph extension

First [set up your development environment](development_environment.md) so you're ready for creating and publishing.

## Create an extension

Creating an extension uses the [Sourcegraph extension creator](https://github.com/sourcegraph/create-extension). It requires no installation and will create all files necessary for development, building, and publishing.

Create a directory for your extension, then use the extension creator:

```shell
mkdir my-extension
cd my-extension
npm init @sourcegraph/extension
```

Follow the prompts. When complete, you'll have the following files:

```shell
├── README.md
├── node_modules
├── package-lock.json
├── package.json
├── src
│   └── my-extension.ts
├── tsconfig.json
└── tslint.json
```

## Extension files

### The src directory and activate function

The extension runtime knows how to "activate" your extension by looking for an `activate` function such as in `src/my-extension.ts`. See the [activation documentation](activation.md) to learn how to control the activation of your extension.

For code layout, a simple extension only needs a single `.ts` file. For larger projects, create multiples files in the `src` directory and these will be bundled into a single file when built for publishing

### README.md

The `README.md` is the content for your extension page in the [extensions registry](https://sourcegraph.com/extensions). See the [Codecov extension](https://sourcegraph.com/extensions/sourcegraph/codecov) for a great example.

### package.json

The `package.json` defines extension configuration and commands required for development and publishing. 

<!--TODO: Ryan: If you're creating your first extension, leave the `package.json` as is or see the [extension configuration documentation](extension_configuration.md).-->

### tslint.json and tsconfig.json

These are configuration files for linting and Typescript compilation and will suffice for most extensions.
