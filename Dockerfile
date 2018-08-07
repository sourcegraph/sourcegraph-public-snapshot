FROM cypress/browsers:chrome65-ff57

ARG NPM_TOKEN

# The base image comes with npm 5.x.x
RUN npm i -g npm@^6.0.0

RUN apt-get update && \
    # Install deps that are usually installed locally or by the CI provider
    # but aren't by buildkite.
    # https://docs.cypress.io/guides/guides/continuous-integration.html#Dependencies
    apt-get install -y xvfb

COPY . /bext

WORKDIR /bext

# Add the .npmrc file
RUN cp ./ci/.npmrc ./.npmrc && \
    # and add the $NPM_TOKEN from args
    sed -i "s/npm_token/${NPM_TOKEN}/g" ./.npmrc

RUN cat .npmrc

RUN npm ci && npm run build
