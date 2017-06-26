# PHP Buildserver

## Needed Environment

- PHP >= 7.0
- [Composer](https://getcomposer.org/)

## Commands

- Install dependencies: `composer install`
- Run tests: `vendor/bin/phpunit`
- Lint: `vendor/bin/phpcs`
- Start over STDIO: `php bin/php-build-server.php`
- Start over TCP: `php bin/php-build-server.php --tcp=127.0.0.1:2088`

## Wiring it up to a local Sourcegraph instance

Start the LS in TCP mode in a terminal with `node lib/language-server --port 2088` or through VS Code, then in a different terminal run

```bash
export LANGSERVER_PHP=tcp://localhost:2088
export LANGSERVER_PHP_BG=tcp://localhost:2089
./dev/start.sh
```

## Developing on the language server

Replace `vendor/felixfbecker/language-server` with a symlink to your local clone.

## Development in VS Code

The launch.json provides a launch configuration for running tests and for running the server.

## Deploying a new version

To release a new version of the language server to packagist.org:  
```bash
git tag vX.X.X
git push --tags
```

To update the language server here:
- run `composer update`
- make sure compilation and tests pass
- commit the updated composer.lock

To deploy it on sourcegraph.com, push `master` to `docker-images/xlang-php`:

    git push -f origin master:docker-images/xlang-php

