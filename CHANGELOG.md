## [0.3.2](https://github.com/stuttgart-things/homerun2-omni-pitcher/compare/v0.3.1...v0.3.2) (2026-03-07)


### Bug Fixes

* adapt to homerun-library API change (RedisConfig struct) ([595cdf5](https://github.com/stuttgart-things/homerun2-omni-pitcher/commit/595cdf568f665de56b16a1319bea412a29e62b7f))
* **deps:** update github.com/stuttgart-things/homerun-library digest to b54ec16 ([ab15762](https://github.com/stuttgart-things/homerun2-omni-pitcher/commit/ab15762de1de297b5b980bf80d2e295df65b45f3))

## [0.3.1](https://github.com/stuttgart-things/homerun2-omni-pitcher/compare/v0.3.0...v0.3.1) (2026-03-07)


### Bug Fixes

* handle errcheck lint failures in auth middleware ([d4388b2](https://github.com/stuttgart-things/homerun2-omni-pitcher/commit/d4388b2e775edd7a71836b17161ec2418130ca92))

# [0.3.0](https://github.com/stuttgart-things/homerun2-omni-pitcher/compare/v0.2.0...v0.3.0) (2026-03-07)


### Features

* switch to log/slog for structured, leveled logging ([b45255a](https://github.com/stuttgart-things/homerun2-omni-pitcher/commit/b45255aaf21506b720dd8da44cb3d23fc916a0fd)), closes [#50](https://github.com/stuttgart-things/homerun2-omni-pitcher/issues/50)

# [0.2.0](https://github.com/stuttgart-things/homerun2-omni-pitcher/compare/v0.1.1...v0.2.0) (2026-03-07)


### Bug Fixes

* use safe defaults for ldflags env vars in ko config ([4157523](https://github.com/stuttgart-things/homerun2-omni-pitcher/commit/4157523600d6c2f7ccad3de08096d61571cc0bb8)), closes [#53](https://github.com/stuttgart-things/homerun2-omni-pitcher/issues/53)


### Features

* embed build info via ldflags (version, commit, date) ([e0d4b06](https://github.com/stuttgart-things/homerun2-omni-pitcher/commit/e0d4b06d85738e47c77c015b8aae58033468549a)), closes [#53](https://github.com/stuttgart-things/homerun2-omni-pitcher/issues/53)

## [0.1.1](https://github.com/stuttgart-things/homerun2-omni-pitcher/compare/v0.1.0...v0.1.1) (2026-03-06)


### Bug Fixes

* accept int or str for redisPort in KCL schema ([#46](https://github.com/stuttgart-things/homerun2-omni-pitcher/issues/46)) ([27b55ef](https://github.com/stuttgart-things/homerun2-omni-pitcher/commit/27b55ef3a8d8b787beb5908008b739a55e77d639))

# [0.1.0](https://github.com/stuttgart-things/homerun2-omni-pitcher/compare/v0.0.0...v0.1.0) (2026-03-06)


### Bug Fixes

* add .goreleaser.yaml for release binary builds ([#42](https://github.com/stuttgart-things/homerun2-omni-pitcher/issues/42)) ([4ee0b8c](https://github.com/stuttgart-things/homerun2-omni-pitcher/commit/4ee0b8c230a93f7b10f9d9405346225e21df7202))
* add Bearer token auth to examples/test-api.sh ([151b501](https://github.com/stuttgart-things/homerun2-omni-pitcher/commit/151b501af800911cd22333889ce9dccc4e27299c)), closes [#10](https://github.com/stuttgart-things/homerun2-omni-pitcher/issues/10)
* **deps:** update github.com/stuttgart-things/homerun-library digest to 7e9992b ([92d6b36](https://github.com/stuttgart-things/homerun2-omni-pitcher/commit/92d6b368bc1f39fec7386b988af70993247947b2))
* remove repositoryUrl from .releaserc to fix release auth ([#39](https://github.com/stuttgart-things/homerun2-omni-pitcher/issues/39)) ([f434bc4](https://github.com/stuttgart-things/homerun2-omni-pitcher/commit/f434bc4401d9f83cd63d85ffcf46ccb8a39612e0))
* remove windows from goreleaser build targets ([#43](https://github.com/stuttgart-things/homerun2-omni-pitcher/issues/43)) ([c8df477](https://github.com/stuttgart-things/homerun2-omni-pitcher/commit/c8df477d4e5d3f4610133c30d54d40eca27a0287))
* revert release workflow to secrets inherit ([#38](https://github.com/stuttgart-things/homerun2-omni-pitcher/issues/38)) ([d1695ea](https://github.com/stuttgart-things/homerun2-omni-pitcher/commit/d1695eaa24f239531c74a3751c415266f1412b6d))
* set Content-Type header before writing response in auth middleware ([1d55c44](https://github.com/stuttgart-things/homerun2-omni-pitcher/commit/1d55c44f4931df552cf67228c230eec792a851ee)), closes [#6](https://github.com/stuttgart-things/homerun2-omni-pitcher/issues/6)
* set kcl-source-dir to kcl in release workflow ([#45](https://github.com/stuttgart-things/homerun2-omni-pitcher/issues/45)) ([d121d65](https://github.com/stuttgart-things/homerun2-omni-pitcher/commit/d121d65f0650728aff76c11aa9b7c5cdcb05e493))


### Features

* add .ko.yaml with explicit build configuration ([3e8de9e](https://github.com/stuttgart-things/homerun2-omni-pitcher/commit/3e8de9e242c88752093101ccb70ef675c5e8c5ca)), closes [#13](https://github.com/stuttgart-things/homerun2-omni-pitcher/issues/13)
* add Backstage catalog, TechDocs, and fix release workflow auth ([#35](https://github.com/stuttgart-things/homerun2-omni-pitcher/issues/35)) ([246c595](https://github.com/stuttgart-things/homerun2-omni-pitcher/commit/246c59568deb4dc57309f138b7204d6aeccd77f7)), closes [#34](https://github.com/stuttgart-things/homerun2-omni-pitcher/issues/34)
* add GitHub Actions CI/CD workflows ([6d22885](https://github.com/stuttgart-things/homerun2-omni-pitcher/commit/6d228857c47daec92dc2c4ca0cb7e84434079f77)), closes [#16](https://github.com/stuttgart-things/homerun2-omni-pitcher/issues/16)
* add HTTPRoute (Gateway API) as alternative to Ingress in KCL ([818068d](https://github.com/stuttgart-things/homerun2-omni-pitcher/commit/818068d31b32f31f43654b5e6c256127727633bb))
* expand Dagger module with Lint, Build, BuildImage, ScanImage ([d6b0baa](https://github.com/stuttgart-things/homerun2-omni-pitcher/commit/d6b0baa4cc08b8b200c4a1491cb59e127e66b33e)), closes [#14](https://github.com/stuttgart-things/homerun2-omni-pitcher/issues/14) [#17](https://github.com/stuttgart-things/homerun2-omni-pitcher/issues/17) [#4](https://github.com/stuttgart-things/homerun2-omni-pitcher/issues/4)
* expand Taskfile with manifest rendering, deploy, release tasks ([babebbc](https://github.com/stuttgart-things/homerun2-omni-pitcher/commit/babebbcdcf8dc6483bd7e97ba074d37601cb724b)), closes [#15](https://github.com/stuttgart-things/homerun2-omni-pitcher/issues/15)
* feat/add-skeleton ([6862da7](https://github.com/stuttgart-things/homerun2-omni-pitcher/commit/6862da7ceadca6f9cc8882d7b2f471cea102a127))
* feat/add-skeleton ([c2091f6](https://github.com/stuttgart-things/homerun2-omni-pitcher/commit/c2091f6791735c888d07ad24c69ad2bc4f411cb6))
* feat/add-skeleton ([f6cab48](https://github.com/stuttgart-things/homerun2-omni-pitcher/commit/f6cab4810f378e2ee2210a014612b11d26a6b02b))
* feat/add-skeleton ([4dd7fc0](https://github.com/stuttgart-things/homerun2-omni-pitcher/commit/4dd7fc0ab9fa1fbf1f10d4e65420777118e42feb))
* feat/add-skeleton ([4f8e7fa](https://github.com/stuttgart-things/homerun2-omni-pitcher/commit/4f8e7fa48e193b02abb204459ed36d3f6ad88cc7))
* feat/add-skeleton ([155b9f4](https://github.com/stuttgart-things/homerun2-omni-pitcher/commit/155b9f4200a313e1788dd54058ebf48ab17f89b4))
* feat/add-skeleton ([8be7255](https://github.com/stuttgart-things/homerun2-omni-pitcher/commit/8be72552fb8a0945de126dea9b4a3c3eab7b46b8))
* feat/add-skeleton ([66f51e8](https://github.com/stuttgart-things/homerun2-omni-pitcher/commit/66f51e8349514da6c60e75059753ba0c9f8b4c97))
* feat/add-skeleton ([1380180](https://github.com/stuttgart-things/homerun2-omni-pitcher/commit/1380180ba59966d8511683a77a978a669ddd3fe7))
* feat/add-skeleton ([359e02d](https://github.com/stuttgart-things/homerun2-omni-pitcher/commit/359e02d1081693412ba72fdd32839c295a1be179))
* feat/add-skeleton ([062db4f](https://github.com/stuttgart-things/homerun2-omni-pitcher/commit/062db4f64800e161e2a3c0ed2242741b88987e53))
* feat/add-skeleton ([80e3128](https://github.com/stuttgart-things/homerun2-omni-pitcher/commit/80e3128739bc7c9f0fc5e381774a0f7499e5f22b))
* feat/add-skeleton ([6ece2ec](https://github.com/stuttgart-things/homerun2-omni-pitcher/commit/6ece2ec9fef606ff9877c360480f67dda424a61c))
* handle Redis errors, load config at startup, add graceful shutdown ([41005b5](https://github.com/stuttgart-things/homerun2-omni-pitcher/commit/41005b520b03046df581168f03f79712e44cb7c7)), closes [#5](https://github.com/stuttgart-things/homerun2-omni-pitcher/issues/5) [#7](https://github.com/stuttgart-things/homerun2-omni-pitcher/issues/7)
* modularize KCL manifests with schema validation ([f03162f](https://github.com/stuttgart-things/homerun2-omni-pitcher/commit/f03162f7e794c20a2c46f33001f648ac9d09eedc)), closes [#11](https://github.com/stuttgart-things/homerun2-omni-pitcher/issues/11) [#3](https://github.com/stuttgart-things/homerun2-omni-pitcher/issues/3)
* remove legacy k8s/ YAML in favor of KCL → Kustomize pipeline ([2ec96b0](https://github.com/stuttgart-things/homerun2-omni-pitcher/commit/2ec96b0a65fd7eed0f43a48f55371c4b4c2666b3)), closes [#12](https://github.com/stuttgart-things/homerun2-omni-pitcher/issues/12)

# Changelog
