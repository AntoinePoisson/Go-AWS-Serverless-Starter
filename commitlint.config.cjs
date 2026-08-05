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
  },
};
