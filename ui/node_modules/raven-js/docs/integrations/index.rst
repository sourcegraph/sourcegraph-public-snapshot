Integrations
============

Integrations extend the functionality of Raven.js to cover common libraries and
environments automatically using simple plugins.

What are plugins?
-----------------

In Raven.js, plugins are little snippets of code to augment functionality
for a specific application/framework. It is highly recommended to checkout
the list of plugins and use what apply to your project.

In order to keep the core small, we have opted to only include the most
basic functionality by default, and you can pick and choose which plugins
are applicable for you.

Why are plugins needed at all?
------------------------------

JavaScript is pretty restrictive when it comes to exception handling, and
there are a lot of things that make it difficult to get relevent
information, so it's important that we inject code and wrap things
magically so we can extract what we need. See :doc:`../usage` for tips
regarding that.

Installation
------------

To install a plugin just include the plugin **after** Raven has been loaded and the Raven global variable is registered. This happens automatically if you install from a CDN with the required plugins in the URL.


.. toctree::
   :maxdepth: 1

   angular
   angular2
   backbone
   ember
   react
   react-native
   vue
