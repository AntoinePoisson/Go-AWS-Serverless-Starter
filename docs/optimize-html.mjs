// Tweaks the HTML `redocly build-docs` writes.
//
// The page is already server-rendered. Redocly still drops a 1.1 MB bundle
// as a blocking <script> in <head>, so the phone parses it before it paints
// anything it already has (~2.6s on a mid-range device). No flag for this,
// the tags come from the template verbatim, so we rewrite them here.
// If Redocly changes its output this script fails instead of shipping a
// slow page again.
//
// Usage: node docs/optimize-html.mjs <file>
//
// TODO: drop this once Redocly lets us defer the bundle.

import { readFileSync, writeFileSync } from 'node:fs';

const file = process.argv[2];

if (!file) {
  console.error('usage: node docs/optimize-html.mjs <file>');
  process.exit(1);
}

// Marker so a second run is a no-op. Wrapping hydrate twice would leave
// the page stuck, the listener never fires again.
const MARKER = '<!-- optimized by docs/optimize-html.mjs -->';

const rewrites = [
  {
    name: 'defer the Redoc bundle',
    // Tag Redocly emits, SRI hash and all. Version is not pinned on purpose.
    from: /(<script src="https:\/\/cdn\.redocly\.com\/redoc\/v[^"]+\/bundles\/redoc\.standalone\.js"[^>]*?)(><\/script>)/,
    to: '$1 defer$2',
  },
  {
    name: 'hydrate once the deferred bundle has run',
    // Deferred scripts run after parse, before DOMContentLoaded, so this
    // still fires as early as it can. A bare call would now run before
    // Redoc exists.
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
