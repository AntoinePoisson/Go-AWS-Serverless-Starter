// Keep this list in sync with changelog-sections in the Release Please config,
// or a commit lints fine here and then vanishes from the changelog.
module.exports = {
  extends: ['@commitlint/config-conventional'],
  rules: {
    'type-enum': [
      2,
      'always',
      [
        'build', // dependencies, packaging, toolchain
        'chore', // hidden from the changelog
        'ci', // workflows, hooks, linters
        'docs', // README and swag annotations
        'feat', // minor bump
        'fix', // patch bump
        'perf',
        'refactor',
        'revert',
        'style', // hidden from the changelog
        'test',
      ],
    ],

    // config-conventional caps every line at 100 chars. The message is where
    // the reasoning lives, so wrapping is left to whoever writes it.
    'header-max-length': [0, 'always', Infinity],
    'body-max-line-length': [0, 'always', Infinity],
    'footer-max-line-length': [0, 'always', Infinity],
  },
};
