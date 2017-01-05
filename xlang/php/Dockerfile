
# Running this container will start a language server that listens for TCP connections on port 2088
# Every connection will be run in a forked child process

# Please note that before building the image, you have to install dependencies with `composer install`

FROM php:7-cli
MAINTAINER Felix Becker <felix.b@outlook.com>

RUN apt-get update \
    && apt-get install -y libxml2 libxml2-dev git unzip \
    && rm -rf /var/lib/apt/lists/*

RUN docker-php-ext-configure pcntl --enable-pcntl
RUN docker-php-ext-install pcntl
COPY ./php.ini /usr/local/etc/php/conf.d/

COPY ./ /srv/phpbs

WORKDIR /srv/phpbs

EXPOSE 2088

CMD ["--tcp-server=0:2088"]

ENTRYPOINT ["php", "bin/php-build-server.php"]
