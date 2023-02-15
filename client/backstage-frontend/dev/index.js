"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
const jsx_runtime_1 = require("react/jsx-runtime");
const dev_utils_1 = require("@backstage/dev-utils");
const plugin_1 = require("../src/plugin");
(0, dev_utils_1.createDevApp)()
    .registerPlugin(plugin_1.williamPlugin)
    .addPage({
    element: (0, jsx_runtime_1.jsx)(plugin_1.WilliamPage, {}),
    title: 'Root Page',
    path: '/william'
})
    .render();
//# sourceMappingURL=index.js.map