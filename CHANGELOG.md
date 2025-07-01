# [2.2.0](https://github.com/M4XGO/terraform-provider-lws/compare/v2.1.0...v2.2.0) (2025-07-01)


### Features

* enable complete GPG signing for Terraform Registry compliance ([6a7dd80](https://github.com/M4XGO/terraform-provider-lws/commit/6a7dd800763ee58cb6b6582737bd7212df25135e))

# [2.1.0](https://github.com/M4XGO/terraform-provider-lws/compare/v2.0.1...v2.1.0) (2025-07-01)


### Features

* add automated GPG setup workflow for Terraform Registry ([8fec6e3](https://github.com/M4XGO/terraform-provider-lws/commit/8fec6e3ef7c12b7254009b2b173ff7863f92d1f1))

## [2.0.1](https://github.com/M4XGO/terraform-provider-lws/compare/v2.0.0...v2.0.1) (2025-07-01)


### Bug Fixes

* enable GPG signing and resolve Terraform Registry validation errors ([47ce61c](https://github.com/M4XGO/terraform-provider-lws/commit/47ce61c1ff31d5abd0999d74821d9c4e62471bdd))

# [2.0.0](https://github.com/M4XGO/terraform-provider-lws/compare/v1.0.3...v2.0.0) (2025-07-01)


### Bug Fixes

* **release:** temporarily disable GPG signing for testing ([e6f902c](https://github.com/M4XGO/terraform-provider-lws/commit/e6f902c4568df6d1656e393108b006b36a1ba9ac))


### BREAKING CHANGES

* **release:** Releases will not be signed until GPG keys are configured in GitHub secrets

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
