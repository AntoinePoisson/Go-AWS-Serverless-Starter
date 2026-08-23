// Post-processes the file `redocly build-docs` writes.
//
// The reference is server-rendered: the HTML already carries the full markup
// and its stylesheet. The Redoc bundle only hydrates it. Redocly still emits
// the bundle as a plain <script> in the <head>, so the browser downloads and
// parses 1.1 MB of JavaScript before painting anything it already holds --
// worth ~2.6s of blocked rendering on a mid-range phone.
//
// Redocly gives no option for this and the tags come from `redocHead` and
// `redocHTML`, which the Handlebars template can only inject verbatim, hence
// this pass. Every rewrite is mandatory: if Redocly changes its output the
// build fails here instead of silently shipping a slow page again.
//
// Usage: node docs/optimize-html.mjs <file>

import { readFileSync, writeFileSync } from 'node:fs';

const file = process.argv[2];

if (!file) {
  console.error('usage: node docs/optimize-html.mjs <file>');
  process.exit(1);
}

// Written into the output so a second run is a no-op instead of wrapping the
// hydration call inside an already-fired DOMContentLoaded listener, which would
// leave the page permanently un-hydrated.
const MARKER = '<!-- optimized by docs/optimize-html.mjs -->';

const rewrites = [
  {
    name: 'defer the Redoc bundle',
    // Matches the tag Redocly builds, SRI hash and all, without pinning the
    // version.
    from: /(<script src="https:\/\/cdn\.redocly\.com\/redoc\/v[^"]+\/bundles\/redoc\.standalone\.js"[^>]*?)(><\/script>)/,
    to: '$1 defer$2',
  },
  {
    name: 'hydrate once the deferred bundle has run',
    // Deferred scripts run after parsing and before DOMContentLoaded, so this
    // listener still fires as early as possible -- but no longer before Redoc
    // exists, which is what a bare call would now do.
    from: /Redoc\.hydrate\(__redoc_state, container\);/,
    to: "document.addEventListener('DOMContentLoaded', function () {\n        Redoc.hydrate(__redoc_state, container);\n      });",
  },
];

let html = readFileSync(file, 'utf8');

if (html.includes(MARKER)) {
  console.log(`docs/optimize-html.mjs: ${file} is already optimized, nothing to do`);
  process.exit(0);
}

for (const { name, from, to } of rewrites) {
  if (!from.test(html)) {
    console.error(`docs/optimize-html.mjs: could not ${name} in ${file}.`);
    console.error('The Redocly output changed. Update the pattern in this file.');
    process.exit(1);
  }

  html = html.replace(from, to);
}

html = html.replace('</head>', `  ${MARKER}\n  </head>`);

writeFileSync(file, html);
console.log(`docs/optimize-html.mjs: ${rewrites.length} rewrites applied to ${file}`);
