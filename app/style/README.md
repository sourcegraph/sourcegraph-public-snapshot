This directory contains [Sass](http://sass-lang.com/) stylesheets.

Include stylesheets in the JavaScript entrypoint you want to use them with.

### Making changes to Bootstrap variables

To override the value of a Bootstrap variable (which are defined in
`node_modules/bootstrap-sass/assets/stylesheets/bootstrap/_variables.scss`),
set it in `_bootstrap_overrides.scss`. Avoid changing files in
`node_modules`, or else they'll get clobbered when we update Bower
dependencies.
