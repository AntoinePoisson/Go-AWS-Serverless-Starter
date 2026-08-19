// The type list below and the changelog-sections of the Release Please config
// are the same list: letting them drift means a commit that lints locally and
// then vanishes from the changelog.
module.exports = {
  extends: ['@commitlint/config-conventional'],
  rules: {
    'type-enum': [
      2,
      'always',
      [
        'build', // dependencies, the Lambda packaging, the toolchain
        'chore', // hidden from the changelog
        'ci', // workflows, hooks, linters
        'docs', // README and the OpenAPI annotations
        'feat', // minor bump
        'fix', // patch bump
        'perf',
        'refactor',
        'revert',
        'style', // hidden from the changelog
        'test',
      ],
    ],

    // config-conventional caps the subject and every body/footer line at 100
    // characters. A commit message is where the reasoning behind a change
    // lives, so nothing here is worth truncating: the caps are off and the
    // wrapping is left to whoever writes the message.
    'header-max-length': [0, 'always', Infinity],
    'body-max-line-length': [0, 'always', Infinity],
    'footer-max-line-length': [0, 'always', Infinity],
  },
};
