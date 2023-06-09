This project was bootstrapped with [eask/cli](https://github.com/emacs-eask/cli).

## How to use?

1. Write and design your package in `your-package.el`
2. Install package dependences if any:

  ```sh
  eask install-deps
  ```

3. Prepare for installation, package it: (it will create package to `dist` folder)

  ```sh
  eask package
  ```

4. Install the built package:

  ```sh
  eask install
  ```

## Compile

You would want to compile your elisp file to check if there are errors:

```sh
eask compile
```

## Cleaning

Simply executes the following:

```sh
eask clean all
```

For more options, see `eask clean --help`!

## Linting

Linting is often optional but recommended to all elisp developers.

with `checkodc`:

```sh
eask lint checkodc
```

with `package-lint`:

```sh
eask lint package  # for pacakge-lint
```

For more options, see `eask lint --help`!

## Testing

Eask supports [ERT](https://www.gnu.org/software/emacs/manual/html_node/ert/index.html)
, [Buttercup](https://github.com/jorgenschaefer/emacs-buttercup)
, [Ecukes](https://github.com/ecukes/ecukes), and more.

For more options, see `eask test --help`!

## Continuous Integration

### GitHub Actions

```sh
eask generate workflow github
```

### CircleCI

```sh
eask generate workflow circle-ci
```

For more options, see `eask generate workflow --help`!

## Learn More

To learn Eask, check out the [Eask documentation](https://github.com/emacs-eask).
