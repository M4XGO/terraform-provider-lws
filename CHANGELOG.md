## [](https://github.com/M4XGO/terraform-provider-lws/compare/v2.1.3...v) (2025-07-03)


### Bug Fixes

* **ci:** use latest commit to create tag and not a new one ([3b8fbf0](https://github.com/M4XGO/terraform-provider-lws/commit/3b8fbf0f2d4701b2daea21f7624936f73fcfa570))

## [2.1.3](https://github.com/M4XGO/terraform-provider-lws/compare/v2.1.2...v2.1.3) (2025-07-03)


### Bug Fixes

* **ci:** use latest commit to create tag and not a new one ([c162703](https://github.com/M4XGO/terraform-provider-lws/commit/c162703b81e41d81389fc113c516056384782772))

## [2.1.2](https://github.com/M4XGO/terraform-provider-lws/compare/v2.1.1...v2.1.2) (2025-07-03)


### Bug Fixes

* use bash script to correctly create release tag ([fe53211](https://github.com/M4XGO/terraform-provider-lws/commit/fe532113d4a9d63c4a79ba102f86fcbcbbb8c4ac))

## [2.1.1](https://github.com/M4XGO/terraform-provider-lws/compare/v2.1.0...v2.1.1) (2025-07-03)


### Features

* add GitHub release configuration to semantic-release workflow ([daf4048](https://github.com/M4XGO/terraform-provider-lws/commit/daf40487fc53dd31913695df2a28e7eddf4630ef))
* enable complete GPG signing for Terraform Registry compliance ([6a7dd80](https://github.com/M4XGO/terraform-provider-lws/commit/6a7dd800763ee58cb6b6582737bd7212df25135e))
* enhance GPG signing configuration for artifact security ([80962a1](https://github.com/M4XGO/terraform-provider-lws/commit/80962a1f24fb46bcb304cec12a5a73ffcae31dfc))
* fix md ([20e721f](https://github.com/M4XGO/terraform-provider-lws/commit/20e721f1901612dbefbbcf23df58c4b34d7dcc92))


### Bug Fixes

* add newlines at the end of multiple files ([230b521](https://github.com/M4XGO/terraform-provider-lws/commit/230b5213b9704e8f1c88cdee35de8511f79dc67d))
* **ci:** add tagMessage to semantic-release config to resolve tag creation error ([54e17cc](https://github.com/M4XGO/terraform-provider-lws/commit/54e17cc390d8b67a78e27dc163d4f747900a6473))
* correct string interpolation in semantic-release configuration ([cb509a5](https://github.com/M4XGO/terraform-provider-lws/commit/cb509a52f6061ba3317b065b7323830bcf8c1d54))
* remove trailing whitespace to resolve gci formatting errors ([2187f09](https://github.com/M4XGO/terraform-provider-lws/commit/2187f09432700522b411135a56c9a466fbb09f40))
* reorder conditions in semantic-release workflow for clarity ([6a97942](https://github.com/M4XGO/terraform-provider-lws/commit/6a979429b009dda749b5c9827e544ae90a7592e5))
* semantic-release-workflow ([aa316da](https://github.com/M4XGO/terraform-provider-lws/commit/aa316daa969b5445742edfe763371e336b72760c))
* semantic-release-workflow ([e193a59](https://github.com/M4XGO/terraform-provider-lws/commit/e193a59068f6f53a58f377724bc969ef1ca8c6ff))
* semantic-release-workflow ([47a1079](https://github.com/M4XGO/terraform-provider-lws/commit/47a1079701da27ff4f6932811c4802cc3e5e474b))
* semantic-release-workflow ([1148402](https://github.com/M4XGO/terraform-provider-lws/commit/114840213e26b72e1c55b7a05164791c0b21d57e))
* semantic-release-workflow ([79d6b4e](https://github.com/M4XGO/terraform-provider-lws/commit/79d6b4e9bf5426c584f8841ff99fb995680c30ee))
* semantic-release-workflow ([438072a](https://github.com/M4XGO/terraform-provider-lws/commit/438072ad3f26970442232235858ca4ecca07678a))
* semantic-release-workflow ([315bdcb](https://github.com/M4XGO/terraform-provider-lws/commit/315bdcb3184cba5065846e92e430f7cd05a59459))
* semantic-release-workflow ([de39289](https://github.com/M4XGO/terraform-provider-lws/commit/de39289ac5b9fbe3ff82cd07be2e928c7b997a24))
* update GPG signing configuration in .goreleaser.yml ([353767a](https://github.com/M4XGO/terraform-provider-lws/commit/353767a4e3575aba94b41f5b589adcb69d4cec64))

## [2.1.0](https://github.com/M4XGO/terraform-provider-lws/compare/v2.0.1...v2.1.0) (2025-07-01)


### Features

* add automated GPG setup workflow for Terraform Registry ([8fec6e3](https://github.com/M4XGO/terraform-provider-lws/commit/8fec6e3ef7c12b7254009b2b173ff7863f92d1f1))

## [2.0.1](https://github.com/M4XGO/terraform-provider-lws/compare/v2.0.0...v2.0.1) (2025-07-01)


### Bug Fixes

* enable GPG signing and resolve Terraform Registry validation errors ([47ce61c](https://github.com/M4XGO/terraform-provider-lws/commit/47ce61c1ff31d5abd0999d74821d9c4e62471bdd))

## [2.0.0](https://github.com/M4XGO/terraform-provider-lws/compare/v1.0.3...v2.0.0) (2025-07-01)


### ⚠ BREAKING CHANGES

* **release:** Releases will not be signed until GPG keys are configured in GitHub secrets

### Bug Fixes

* **release:** temporarily disable GPG signing for testing ([e6f902c](https://github.com/M4XGO/terraform-provider-lws/commit/e6f902c4568df6d1656e393108b006b36a1ba9ac))

## [1.0.3](https://github.com/M4XGO/terraform-provider-lws/compare/v1.0.2...v1.0.3) (2025-07-01)


### Bug Fixes

* **ci:** remove standalone gosec job and integrate security scanning into golangci-lint ([3f787e0](https://github.com/M4XGO/terraform-provider-lws/commit/3f787e0807c1812a584379cd03c90a7ac3d26277))

## [1.0.2](https://github.com/M4XGO/terraform-provider-lws/compare/v1.0.1...v1.0.2) (2025-07-01)


### Bug Fixes

* **ci:** improve gosec installation reliability ([422df97](https://github.com/M4XGO/terraform-provider-lws/commit/422df97d953508781195a429d6840ed41f7c3f75))

## [1.0.1](https://github.com/M4XGO/terraform-provider-lws/compare/v1.0.0...v1.0.1) (2025-07-01)


### Bug Fixes

* **ci:** resolve go generate, gosec, and terraform validation issues ([5f88ff9](https://github.com/M4XGO/terraform-provider-lws/commit/5f88ff95433d6cc46c8a81cf74cfb6b1dc1946f6))
* **ci:** resolve gosec installation and goreleaser version compatibility issues ([602774c](https://github.com/M4XGO/terraform-provider-lws/commit/602774c00362f07a96a5da3e733cc98ebf65f3f0))
* **ci:** resolve npm cache and gosec installation issues ([8083609](https://github.com/M4XGO/terraform-provider-lws/commit/80836099ca6392569b7741567b9b72b60905c21f))

## [1.0.0](https://github.com/M4XGO/terraform-provider-lws/compare/427748c4e9c4791056d2bb68d202b960c245a19f...v1.0.0) (2025-07-01)


### Features

* initial terraform provider for LWS with DNS record management ([427748c](https://github.com/M4XGO/terraform-provider-lws/commit/427748c4e9c4791056d2bb68d202b960c245a19f))


### Bug Fixes

* update .gitignore to include mykey-private.asc ([149e5ed](https://github.com/M4XGO/terraform-provider-lws/commit/149e5ede1f395c9f207a2e0eb992545743c9db6b))

