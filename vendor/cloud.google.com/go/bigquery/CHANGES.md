# Changes




## [1.60.0](https://github.com/googleapis/google-cloud-go/compare/bigquery/v1.59.1...bigquery/v1.60.0) (2024-03-27)


### Features

* **bigquery/analyticshub:** Support selective sharing on data clean room Listings ([a3bb7c0](https://github.com/googleapis/google-cloud-go/commit/a3bb7c07ba570f26c6eb073ab3275487784547d0))
* **bigquery/datatransfer:** Add UnenrollDataSources API which gives users a programmatic way to unenroll data sources ([a86aa8e](https://github.com/googleapis/google-cloud-go/commit/a86aa8e962b77d152ee6cdd433ad94967150ef21))
* **bigquery/storage:** Add the RANGE type to the google.cloud.bigquery.storage.v1.TableFieldSchema ([0195fe9](https://github.com/googleapis/google-cloud-go/commit/0195fe9292274ff9d86c71079a8e96ed2e5f9331))


### Bug Fixes

* **bigquery/storage/managedwriter:** Fix flowcontrol refund on error ([#9649](https://github.com/googleapis/google-cloud-go/issues/9649)) ([a07bf1d](https://github.com/googleapis/google-cloud-go/commit/a07bf1d463416d8a4bbfbbd7e720653962fb27d5))
* **bigquery/storage/managedwriter:** Retry improvements ([#9642](https://github.com/googleapis/google-cloud-go/issues/9642)) ([48a9258](https://github.com/googleapis/google-cloud-go/commit/48a9258954b9be40d74656dc12fe46f2bbc19bda))
* **bigquery:** Update protobuf dep to v1.33.0 ([30b038d](https://github.com/googleapis/google-cloud-go/commit/30b038d8cac0b8cd5dd4761c87f3f298760dd33a))


### Documentation

* **bigquery/datatransfer:** Update unenrollDataSources API documentation ([#9449](https://github.com/googleapis/google-cloud-go/issues/9449)) ([da644cc](https://github.com/googleapis/google-cloud-go/commit/da644ccc6f154d41b8bedf21f4c4cc8aa9ce41bf))
* **bigquery/storage:** Mark BigQueryWrite v1beta2 as deprecated ([d130d86](https://github.com/googleapis/google-cloud-go/commit/d130d861f55d137a2803340c2e11da3589669cb8))

## [1.59.1](https://github.com/googleapis/google-cloud-go/compare/bigquery/v1.59.0...bigquery/v1.59.1) (2024-02-12)


### Bug Fixes

* **bigquery:** Align return time.Time values to UTC ([#9411](https://github.com/googleapis/google-cloud-go/issues/9411)) ([4ac005d](https://github.com/googleapis/google-cloud-go/commit/4ac005d77f1e19981c7e1ab1cebec1302e0840ca))

## [1.59.0](https://github.com/googleapis/google-cloud-go/compare/bigquery/v1.58.0...bigquery/v1.59.0) (2024-02-06)


### Features

* **bigquery:** Add ExportDataStatstics to QueryStatistics ([#9371](https://github.com/googleapis/google-cloud-go/issues/9371)) ([261c8d9](https://github.com/googleapis/google-cloud-go/commit/261c8d944b53ac8953ea7d771c4bb50e4078d508))
* **bigquery:** Switch all timestamp representations to int64 usec ([#9368](https://github.com/googleapis/google-cloud-go/issues/9368)) ([8c1fb7d](https://github.com/googleapis/google-cloud-go/commit/8c1fb7d4728ebc3b21cb0d601952966dca9cd1e8))


### Bug Fixes

* **bigquery/storage/managedwriter:** Resolve data races ([#9360](https://github.com/googleapis/google-cloud-go/issues/9360)) ([fa31ec0](https://github.com/googleapis/google-cloud-go/commit/fa31ec0c0f04302a9713a9b1d3228bda2ba135c6))
* **bigquery:** Enable universe domain resolution options ([fd1d569](https://github.com/googleapis/google-cloud-go/commit/fd1d56930fa8a747be35a224611f4797b8aeb698))
* **bigquery:** Support more timestamp formats for query param ([#9236](https://github.com/googleapis/google-cloud-go/issues/9236)) ([cc98509](https://github.com/googleapis/google-cloud-go/commit/cc98509fc7961e3d3619b837d13e69f9621386e8)), refs [#9221](https://github.com/googleapis/google-cloud-go/issues/9221)

## [1.58.0](https://github.com/googleapis/google-cloud-go/compare/bigquery/v1.57.1...bigquery/v1.58.0) (2024-01-18)


### Features

* **bigquery/storage/managedwriter:** Improve schema comparison stability ([#9241](https://github.com/googleapis/google-cloud-go/issues/9241)) ([faccb68](https://github.com/googleapis/google-cloud-go/commit/faccb68deeb18ad51875042242a1141285448bbb))
* **bigquery/storage:** Add ability to request compressed ReadRowsResponse rows ([c3f1174](https://github.com/googleapis/google-cloud-go/commit/c3f1174dc29d1c00d514a69590bd83f9b08a60d1))
* **bigquery:** Add DataGovernanceType to routines ([#8990](https://github.com/googleapis/google-cloud-go/issues/8990)) ([57491ae](https://github.com/googleapis/google-cloud-go/commit/57491ae3b7bbcee2abe1064dfb4f18cb3a511265))
* **bigquery:** Add schema support for RANGE type ([#9050](https://github.com/googleapis/google-cloud-go/issues/9050)) ([477ccee](https://github.com/googleapis/google-cloud-go/commit/477ccee8eb210a0e0f9c71513318e0fbbd08928c))
* **bigquery:** Add support for AllowNonIncrementalDefinition and Staleness on MaterializedView ([#8673](https://github.com/googleapis/google-cloud-go/issues/8673)) ([6ec2bb2](https://github.com/googleapis/google-cloud-go/commit/6ec2bb218335acf791af859852fc6c8e9bea7a08))
* **bigquery:** Add table resource tags support ([#9084](https://github.com/googleapis/google-cloud-go/issues/9084)) ([3569cc2](https://github.com/googleapis/google-cloud-go/commit/3569cc27b3357c3f62ccfae1670a1bf6383dee9a))
* **bigquery:** Expose query id on row iterator if available ([#9224](https://github.com/googleapis/google-cloud-go/issues/9224)) ([bbff8ac](https://github.com/googleapis/google-cloud-go/commit/bbff8ac817fb95af219c588bdadc0ca1784c5c0c))


### Documentation

* **bigquery/storage/managedwriter:** Add documentation for ingesting JSON to a ManagedStream ([#9137](https://github.com/googleapis/google-cloud-go/issues/9137)) ([cae54ed](https://github.com/googleapis/google-cloud-go/commit/cae54ed31548e6f17f3b6bc217bb117723f04e15))
* **bigquery:** Update package docs to show query parameterization ([#8965](https://github.com/googleapis/google-cloud-go/issues/8965)) ([cc76ab3](https://github.com/googleapis/google-cloud-go/commit/cc76ab35c25a672ed3255c0b26a99bef2cb81868))

## [1.57.1](https://github.com/googleapis/google-cloud-go/compare/bigquery/v1.57.0...bigquery/v1.57.1) (2023-11-01)


### Bug Fixes

* **bigquery:** Correct deps issue ([#8958](https://github.com/googleapis/google-cloud-go/issues/8958)) ([abd6ca7](https://github.com/googleapis/google-cloud-go/commit/abd6ca7e4d59595abe6ffb2abf738b445955d0ce))

## [1.57.0](https://github.com/googleapis/google-cloud-go/compare/bigquery/v1.56.0...bigquery/v1.57.0) (2023-10-30)


### Features

* **bigquery/biglake:** Promote to GA ([e864fbc](https://github.com/googleapis/google-cloud-go/commit/e864fbcbc4f0a49dfdb04850b07451074c57edc8))
* **bigquery/storage/managedwriter:** Support default value controls ([#8686](https://github.com/googleapis/google-cloud-go/issues/8686)) ([dfa8e22](https://github.com/googleapis/google-cloud-go/commit/dfa8e22edf560211ae2a2ebf1f9a23b86887c7be))
* **bigquery:** Expose Apache Arrow data through ArrowIterator  ([#8506](https://github.com/googleapis/google-cloud-go/issues/8506)) ([c8e7692](https://github.com/googleapis/google-cloud-go/commit/c8e76923621b379fb7deb6dfb944011af1d980bd)), refs [#8100](https://github.com/googleapis/google-cloud-go/issues/8100)
* **bigquery:** Introduce query preview features ([#8653](https://github.com/googleapis/google-cloud-go/issues/8653)) ([f29683b](https://github.com/googleapis/google-cloud-go/commit/f29683bcd06567e4fc2d404f53bedbea5b5f0f90))


### Bug Fixes

* **bigquery:** Handle storage read api Recv call errors ([#8666](https://github.com/googleapis/google-cloud-go/issues/8666)) ([c73963f](https://github.com/googleapis/google-cloud-go/commit/c73963f64ef667daa8a33a5a4cc2156818fc6914))
* **bigquery:** Update golang.org/x/net to v0.17.0 ([174da47](https://github.com/googleapis/google-cloud-go/commit/174da47254fefb12921bbfc65b7829a453af6f5d))
* **bigquery:** Update grpc-go to v1.56.3 ([343cea8](https://github.com/googleapis/google-cloud-go/commit/343cea8c43b1e31ae21ad50ad31d3b0b60143f8c))
* **bigquery:** Update grpc-go to v1.59.0 ([81a97b0](https://github.com/googleapis/google-cloud-go/commit/81a97b06cb28b25432e4ece595c55a9857e960b7))

## [1.56.0](https://github.com/googleapis/google-cloud-go/compare/bigquery/v1.55.0...bigquery/v1.56.0) (2023-10-05)


### Features

* **bigquery/analyticshub:** Add Subscription resource and RPCs ([#8612](https://github.com/googleapis/google-cloud-go/issues/8612)) ([9992249](https://github.com/googleapis/google-cloud-go/commit/999224951d586513bd382483326b455c953a14f3))
* **bigquery:** Add external dataset reference ([#8545](https://github.com/googleapis/google-cloud-go/issues/8545)) ([1001acf](https://github.com/googleapis/google-cloud-go/commit/1001acf6ac894a7d9945b7d204ad55aa7b162909))
* **bigquery:** Add media options to LoadConfig ([#8640](https://github.com/googleapis/google-cloud-go/issues/8640)) ([62baf56](https://github.com/googleapis/google-cloud-go/commit/62baf569c052a8c565710533ba50a1f5df8176dd))


### Bug Fixes

* **bigquery/storage/managedwriter:** Automatic retry for multiplex test ([#8601](https://github.com/googleapis/google-cloud-go/issues/8601)) ([6ef1945](https://github.com/googleapis/google-cloud-go/commit/6ef1945a837430afcd195affd661768817cdc40b))
* **bigquery:** Dependency detection on proto conversion ([#8566](https://github.com/googleapis/google-cloud-go/issues/8566)) ([763ab5d](https://github.com/googleapis/google-cloud-go/commit/763ab5dbd466fd0450dcd46e38c75e87cbf6c49c))


### Documentation

* **bigquery/datatransfer:** Update transferConfig.name description to indicate that it supports both formats ([0449518](https://github.com/googleapis/google-cloud-go/commit/0449518f8396cc0280c0f3303c103edcee34016b))

## [1.55.0](https://github.com/googleapis/google-cloud-go/compare/bigquery/v1.54.0...bigquery/v1.55.0) (2023-09-11)


### Features

* **bigquery/datapolicies:** Support using custom UDF in the data policy ([20725c8](https://github.com/googleapis/google-cloud-go/commit/20725c86c970ad24efa18c056fc3aa71dc3a4f03))
* **bigquery:** Set PreferredMinStreamCount when creating read session ([#8476](https://github.com/googleapis/google-cloud-go/issues/8476)) ([22e095a](https://github.com/googleapis/google-cloud-go/commit/22e095adea77a2bc3900f5d5c57715d4a9ed4fcb)), refs [#8432](https://github.com/googleapis/google-cloud-go/issues/8432)


### Bug Fixes

* **bigquery/storage/managedwriter:** Address possible deadlocks ([#8507](https://github.com/googleapis/google-cloud-go/issues/8507)) ([48b08bf](https://github.com/googleapis/google-cloud-go/commit/48b08bf47494052a2652a3ab519de76fddc74c8a))
* **bigquery:** Field descriptor proto name should not be lowercase ([#8495](https://github.com/googleapis/google-cloud-go/issues/8495)) ([4287e4b](https://github.com/googleapis/google-cloud-go/commit/4287e4bed40e69203ddb55bd754e3adbf269ba9e))
* **bigquery:** Value for datasetID on foreign keys ([#8447](https://github.com/googleapis/google-cloud-go/issues/8447)) ([fa6e827](https://github.com/googleapis/google-cloud-go/commit/fa6e827a5997fb7abe97375ea6abb7d2fc71047d)), refs [#8442](https://github.com/googleapis/google-cloud-go/issues/8442)

## [1.54.0](https://github.com/googleapis/google-cloud-go/compare/bigquery/v1.53.0...bigquery/v1.54.0) (2023-08-17)


### Features

* **bigquery/biglake:** Start generating apiv1 and apiv1alpha1 ([#8433](https://github.com/googleapis/google-cloud-go/issues/8433)) ([916f779](https://github.com/googleapis/google-cloud-go/commit/916f779a3513b2532d2e0fbd15bfd381ae3efeb5))
* **bigquery/datatransfer:** Add EncryptionConfiguration to TransferConfig ([e3f8c89](https://github.com/googleapis/google-cloud-go/commit/e3f8c89429a207c05fee36d5d93efe76f9e29efe))
* **bigquery/storage/managedwriter:** Refine connection metrics ([#8324](https://github.com/googleapis/google-cloud-go/issues/8324)) ([54fcf36](https://github.com/googleapis/google-cloud-go/commit/54fcf36fe7e26d6e3d11deec19f56e92ceb87d34))
* **bigquery/storage:** Add beta2 deprecation message ([b3dbdde](https://github.com/googleapis/google-cloud-go/commit/b3dbdde48ddfa215c3c3bb110e0051fd8158f451))
* **bigquery/storage:** Add default_missing_value_interpretation field; indicate KMS_SERVICE_ERROR is retryable ([b3dbdde](https://github.com/googleapis/google-cloud-go/commit/b3dbdde48ddfa215c3c3bb110e0051fd8158f451))
* **bigquery:** Add ProjectID to JobIDConfig ([#8405](https://github.com/googleapis/google-cloud-go/issues/8405)) ([7fafd80](https://github.com/googleapis/google-cloud-go/commit/7fafd805327a0734e1602a597c2ded09a860f405))


### Bug Fixes

* **bigquery/datapolicies:** Sync the new PredefinedEpxression types to the client library ([fcb41cc](https://github.com/googleapis/google-cloud-go/commit/fcb41cc1d2435452ee78314c1b0362e3f21ae637))
* **bigquery:** Use JobFromProject for storage iterator ([#8372](https://github.com/googleapis/google-cloud-go/issues/8372)) ([bfbf18c](https://github.com/googleapis/google-cloud-go/commit/bfbf18cb373cd6a88a22eec8db7033f6678df755))


### Documentation

* **bigquery/storage/managedwriter:** Augment context usage in package docs ([#8334](https://github.com/googleapis/google-cloud-go/issues/8334)) ([b30a7f4](https://github.com/googleapis/google-cloud-go/commit/b30a7f4a520c358eb71e1378b98e7413c2ea7f81))
* **bigquery:** Improve RowIterator docs and out of process pagination with Storage API ([#8419](https://github.com/googleapis/google-cloud-go/issues/8419)) ([40f8f80](https://github.com/googleapis/google-cloud-go/commit/40f8f809d9ad1e6d2d0f6207cdf38428c306cb66))

## [1.53.0](https://github.com/googleapis/google-cloud-go/compare/bigquery/v1.52.0...bigquery/v1.53.0) (2023-07-24)


### Features

* **bigquery/analyticshub:** Promote to GA ([130c571](https://github.com/googleapis/google-cloud-go/commit/130c5713dcbac7f670cb92ea113dd53d8029c960))
* **bigquery/connection:** Add support for Salesforce connections, which are usable only by allowlisted partners ([bac978a](https://github.com/googleapis/google-cloud-go/commit/bac978ace43bb58db7c0b1475e41c8fdf8c49a29))
* **bigquery/datapolicies:** Promote to GA ([130c571](https://github.com/googleapis/google-cloud-go/commit/130c5713dcbac7f670cb92ea113dd53d8029c960))
* **bigquery/storage:** Add ResourceExhausted to retryable error for Write API unary calls ([#8214](https://github.com/googleapis/google-cloud-go/issues/8214)) ([8ff13bf](https://github.com/googleapis/google-cloud-go/commit/8ff13bf87397ad524019268c1146e44f3c1cd0e6))


### Bug Fixes

* **bigquery/storage/managedwriter:** Context refactoring ([#8275](https://github.com/googleapis/google-cloud-go/issues/8275)) ([c4104ea](https://github.com/googleapis/google-cloud-go/commit/c4104eaab0d7291c15aba37b78e71ce3cbb9f77a))

## [1.52.0](https://github.com/googleapis/google-cloud-go/compare/bigquery/v1.51.2...bigquery/v1.52.0) (2023-06-23)


### Features

* **bigquery/storage:** Add estimated physical file sizes to ReadAPI v1 ([94ea341](https://github.com/googleapis/google-cloud-go/commit/94ea3410e233db6040a7cb0a931948f1e3bb4c9a))
* **bigquery/storage:** Add table sampling to ReadAPI v1 ([ca94e27](https://github.com/googleapis/google-cloud-go/commit/ca94e2724f9e2610b46aefd0a3b5ddc06102e91b))
* **bigquery:** Support for tables primary and foreign keys ([#8055](https://github.com/googleapis/google-cloud-go/issues/8055)) ([93d6a1a](https://github.com/googleapis/google-cloud-go/commit/93d6a1a1a3bde8d3519acc2b7e77bf8b7ba1678a))
* **bigquery:** Update all direct dependencies ([b340d03](https://github.com/googleapis/google-cloud-go/commit/b340d030f2b52a4ce48846ce63984b28583abde6))


### Bug Fixes

* **bigquery/storage/managedwriter:** Correct reconnection logic ([#8164](https://github.com/googleapis/google-cloud-go/issues/8164)) ([a67d53d](https://github.com/googleapis/google-cloud-go/commit/a67d53ddf13b7d382d4c7856cafb068919021912))
* **bigquery:** REST query UpdateMask bug ([df52820](https://github.com/googleapis/google-cloud-go/commit/df52820b0e7721954809a8aa8700b93c5662dc9b))
* **bigquery:** RowIterator.Schema not filled when using Storage Read API ([#7671](https://github.com/googleapis/google-cloud-go/issues/7671)) ([31040e8](https://github.com/googleapis/google-cloud-go/commit/31040e8a7989b143c0c3c3f3e31c4a9dfbba8094))

## [1.51.2](https://github.com/googleapis/google-cloud-go/compare/bigquery/v1.51.1...bigquery/v1.51.2) (2023-05-08)


### Bug Fixes

* **bigquery:** Update grpc to v1.55.0 ([1147ce0](https://github.com/googleapis/google-cloud-go/commit/1147ce02a990276ca4f8ab7a1ab65c14da4450ef))

## [1.51.1](https://github.com/googleapis/google-cloud-go/compare/bigquery/v1.51.0...bigquery/v1.51.1) (2023-05-04)


### Bug Fixes

* **bigquery/storage/managedwriter:** Schema evolution improvements ([#7838](https://github.com/googleapis/google-cloud-go/issues/7838)) ([49a1621](https://github.com/googleapis/google-cloud-go/commit/49a1621254cc9cc16cbcc2ca779979edfd2565ab))
* **bigquery:** Upgrade arrow to v12 to fix build error ([#7870](https://github.com/googleapis/google-cloud-go/issues/7870)) ([402e365](https://github.com/googleapis/google-cloud-go/commit/402e3651c18b9e5b73529e9a8f9bdeac71323875))


### Documentation

* **bigquery/storage/managedwriter:** Improve tuning guidance ([#7848](https://github.com/googleapis/google-cloud-go/issues/7848)) ([ca2184c](https://github.com/googleapis/google-cloud-go/commit/ca2184c8e7094294479665ca02ae9b29c21b09aa))

## [1.51.0](https://github.com/googleapis/google-cloud-go/compare/bigquery/v1.50.0...bigquery/v1.51.0) (2023-04-17)


### Features

* **bigquery/storage/managedwriter:** Expose connection multiplexing as experimental ([#7673](https://github.com/googleapis/google-cloud-go/issues/7673)) ([3b8bfb4](https://github.com/googleapis/google-cloud-go/commit/3b8bfb44a237fc0696e77a47c5a8305caa5ec64b))
* **bigquery:** Add Iceberg as DataFormat option. ([#7155](https://github.com/googleapis/google-cloud-go/issues/7155)) ([7a9e211](https://github.com/googleapis/google-cloud-go/commit/7a9e21121ca8debffe11fe16d1b96b0d969e11ac))
* **bigquery:** Expose dataset MaxTimeTravelHours ([#7706](https://github.com/googleapis/google-cloud-go/issues/7706)) ([22a666c](https://github.com/googleapis/google-cloud-go/commit/22a666c766fa4565ee120574709b345e7fc37c5f))


### Bug Fixes

* **bigquery:** Respect context during query execution ([#7693](https://github.com/googleapis/google-cloud-go/issues/7693)) ([56772f5](https://github.com/googleapis/google-cloud-go/commit/56772f50217574e6776efeefef767c16e374e593))

## [1.50.0](https://github.com/googleapis/google-cloud-go/compare/bigquery/v1.49.0...bigquery/v1.50.0) (2023-04-03)


### Features

* **bigquery/connection:** Add spark connection properties type ([#7570](https://github.com/googleapis/google-cloud-go/issues/7570)) ([499b489](https://github.com/googleapis/google-cloud-go/commit/499b489d8d6bc8db203c864db97f1462bbeff3d2))
* **bigquery/migration:** Add request_source field and update formatting ([#7586](https://github.com/googleapis/google-cloud-go/issues/7586)) ([c967961](https://github.com/googleapis/google-cloud-go/commit/c967961ed95750e173af0193ec8d0974471f43ff))
* **bigquery/reservation:** Add edition/autoscale related fields ([#7608](https://github.com/googleapis/google-cloud-go/issues/7608)) ([2b7bb66](https://github.com/googleapis/google-cloud-go/commit/2b7bb662eb00671b8ee933766f4254f897131a7c))
* **bigquery/storage/managedwriter:** Decouple connections and writers ([#7314](https://github.com/googleapis/google-cloud-go/issues/7314)) ([7d085b4](https://github.com/googleapis/google-cloud-go/commit/7d085b4b25a29ff1a81164409fc68b8bcb5eacc4))
* **bigquery/storage/managedwriter:** Introduce location routing header ([#7663](https://github.com/googleapis/google-cloud-go/issues/7663)) ([cf06802](https://github.com/googleapis/google-cloud-go/commit/cf068024f1066ee391191066039d7ba2668dd3f4))


### Bug Fixes

* **bigquery/storage/managedwriter:** Fix option propagation ([#7669](https://github.com/googleapis/google-cloud-go/issues/7669)) ([f684e16](https://github.com/googleapis/google-cloud-go/commit/f684e1610c51311c597763b5d1447c178173940a))


### Documentation

* **bigquery/reservation:** Mention that some fields are deprecated ([597ea0f](https://github.com/googleapis/google-cloud-go/commit/597ea0fe09bcea04e884dffe78add850edb2120d))

## [1.49.0](https://github.com/googleapis/google-cloud-go/compare/bigquery/v1.48.0...bigquery/v1.49.0) (2023-03-16)


### Features

* **bigquery:** Add support for storage billing model ([#7510](https://github.com/googleapis/google-cloud-go/issues/7510)) ([0132ca9](https://github.com/googleapis/google-cloud-go/commit/0132ca9e43f979d0e164b31a1fde203694311b43)), refs [#6978](https://github.com/googleapis/google-cloud-go/issues/6978)
* **bigquery:** Update iam and longrunning deps ([91a1f78](https://github.com/googleapis/google-cloud-go/commit/91a1f784a109da70f63b96414bba8a9b4254cddd))

## [1.48.0](https://github.com/googleapis/google-cloud-go/compare/bigquery/v1.47.0...bigquery/v1.48.0) (2023-03-01)


### Features

* **bigquery/connection:** Add cloud spanner connection properties - serverless analytics ([#7487](https://github.com/googleapis/google-cloud-go/issues/7487)) ([14771b1](https://github.com/googleapis/google-cloud-go/commit/14771b146b7add234183d0884ac822b2f05af0e5))
* **bigquery/storage/managedwriter:** Mark managedwriter as GA ([#6804](https://github.com/googleapis/google-cloud-go/issues/6804)) ([3d3eeda](https://github.com/googleapis/google-cloud-go/commit/3d3eedacbaa209ac63f9ada17074620585d03726))

## [1.47.0](https://github.com/googleapis/google-cloud-go/compare/bigquery/v1.46.0...bigquery/v1.47.0) (2023-02-22)


### Features

* **bigquery/storage:** Add default_value_expression to TableFieldSchema ([#7400](https://github.com/googleapis/google-cloud-go/issues/7400)) ([1244b3f](https://github.com/googleapis/google-cloud-go/commit/1244b3f63338ffe7af9051fd368b6b81ca63ff37))
* **bigquery:** Add support for session in load jobs ([#7418](https://github.com/googleapis/google-cloud-go/issues/7418)) ([f9ff2ca](https://github.com/googleapis/google-cloud-go/commit/f9ff2ca3a01736b049873ff43005626e8c06e6b2))


### Bug Fixes

* **bigquery:** Avoid double-channel-close during context cancellation ([#7467](https://github.com/googleapis/google-cloud-go/issues/7467)) ([ca4b2ef](https://github.com/googleapis/google-cloud-go/commit/ca4b2efbda670566f2cfed4eacd559f31cf460ff))

## [1.46.0](https://github.com/googleapis/google-cloud-go/compare/bigquery/v1.45.0...bigquery/v1.46.0) (2023-02-06)


### Features

* **bigquery:** Add dataset/table collation ([#7235](https://github.com/googleapis/google-cloud-go/issues/7235)) ([9f7bbeb](https://github.com/googleapis/google-cloud-go/commit/9f7bbeb466bd7572544c4178a33370a25b5f1496))
* **bigquery:** Use storage api for query jobs ([#6822](https://github.com/googleapis/google-cloud-go/issues/6822)) ([26c04f4](https://github.com/googleapis/google-cloud-go/commit/26c04f4cd5083b4aa3c219500572d3af2f291645))


### Bug Fixes

* **bigquery:** Create/update an isolated dataset for collation feature ([#7256](https://github.com/googleapis/google-cloud-go/issues/7256)) ([b371558](https://github.com/googleapis/google-cloud-go/commit/b3715585aa6892fc41a29027694c72f31390441a))
* **bigquery:** Fetch dst table for jobs when readings with Storage API ([#7325](https://github.com/googleapis/google-cloud-go/issues/7325)) ([0bf80d7](https://github.com/googleapis/google-cloud-go/commit/0bf80d72a893755adefdead900e8990ed53d9627)), refs [#7322](https://github.com/googleapis/google-cloud-go/issues/7322)

## [1.45.0](https://github.com/googleapis/google-cloud-go/compare/bigquery/v1.44.0...bigquery/v1.45.0) (2023-01-05)


### Features

* **bigquery/datapolicies:** Start generating apiv1 ([#7204](https://github.com/googleapis/google-cloud-go/issues/7204)) ([fd71cba](https://github.com/googleapis/google-cloud-go/commit/fd71cba7b6d5a015dcdb24b9eacc7fae1aa54c89))
* **bigquery/datatransfer:** Add location methods ([06a54a1](https://github.com/googleapis/google-cloud-go/commit/06a54a16a5866cce966547c51e203b9e09a25bc0))
* **bigquery:** Add REST client ([06a54a1](https://github.com/googleapis/google-cloud-go/commit/06a54a16a5866cce966547c51e203b9e09a25bc0))
* **bigquery:** Rewrite signatures and type in terms of new location ([620e6d8](https://github.com/googleapis/google-cloud-go/commit/620e6d828ad8641663ae351bfccfe46281e817ad))

## [1.44.0](https://github.com/googleapis/google-cloud-go/compare/bigquery/v1.43.0...bigquery/v1.44.0) (2022-11-30)


### Features

* **bigquery/datatransfer:** added Location API methods docs: updated comments ([22ec3e3](https://github.com/googleapis/google-cloud-go/commit/22ec3e3e727f8c0232059a5d31bccd12b7b5034c))
* **bigquery/storage:** add missing_value_interpretations to AppendRowsRequest ([2a0b1ae](https://github.com/googleapis/google-cloud-go/commit/2a0b1aeb1683222e6aa5c876cb945845c00cef79))
* **bigquery:** Add default partition expiration to Dataset ([#7096](https://github.com/googleapis/google-cloud-go/issues/7096)) ([601c77a](https://github.com/googleapis/google-cloud-go/commit/601c77a69a27b5f13ebb4508f8222a98c8a904bc)), refs [#7021](https://github.com/googleapis/google-cloud-go/issues/7021)
* **bigquery:** Remove code for reservation/apiv1beta1 ([#7010](https://github.com/googleapis/google-cloud-go/issues/7010)) ([451acc1](https://github.com/googleapis/google-cloud-go/commit/451acc1bfc04cc600ab3c1f50f5494e609e65ce2))
* **bigquery:** Start generating proto stubs ([#7026](https://github.com/googleapis/google-cloud-go/issues/7026)) ([debc4c7](https://github.com/googleapis/google-cloud-go/commit/debc4c70786fece5d04d8cad9e9211c55a0a692f))
* **bigquery:** Widen retry predicate ([#6976](https://github.com/googleapis/google-cloud-go/issues/6976)) ([753b751](https://github.com/googleapis/google-cloud-go/commit/753b75139f4b9e8593db5d45d8ab1e0cc8969350))


### Bug Fixes

* **bigquery/storage/managedwriter:** Remove old header routing ([#6960](https://github.com/googleapis/google-cloud-go/issues/6960)) ([434b407](https://github.com/googleapis/google-cloud-go/commit/434b407f4ba66247cb0a15288a2de8e76b691605))


### Documentation

* **bigquery/storage:** remove stale header guidance for AppendRows ([9c5d6c8](https://github.com/googleapis/google-cloud-go/commit/9c5d6c857b9deece4663d37fc6c834fd758b98ca))

## [1.43.0](https://github.com/googleapis/google-cloud-go/compare/bigquery/v1.42.0...bigquery/v1.43.0) (2022-10-24)


### Features

* **bigquery/analyticshub:** rename nodejs analyticshub library package name ([52dddd1](https://github.com/googleapis/google-cloud-go/commit/52dddd1ed89fbe77e1859311c3b993a77a82bfc7))
* **bigquery/storage/managedwriter:** Enable field name indirection ([#6247](https://github.com/googleapis/google-cloud-go/issues/6247)) ([1969273](https://github.com/googleapis/google-cloud-go/commit/19692735b0fbafa176d0315bda923528e1eedf6d))
* **bigquery/storage/managedwriter:** Retry on FailedPrecondition ([#6761](https://github.com/googleapis/google-cloud-go/issues/6761)) ([d1a444d](https://github.com/googleapis/google-cloud-go/commit/d1a444d769c9578b586bef608d343b4b0abd3658))
* **bigquery/storage/managedwriter:** Support append retries ([#6695](https://github.com/googleapis/google-cloud-go/issues/6695)) ([6ae9c67](https://github.com/googleapis/google-cloud-go/commit/6ae9c670a11d80b34872cb05fda933303b81851d))
* **bigquery/storage/managedwriter:** Switch to opt-in retry ([#6765](https://github.com/googleapis/google-cloud-go/issues/6765)) ([a3e97a6](https://github.com/googleapis/google-cloud-go/commit/a3e97a6f15ad1989ef815b9bd5838192f9f226f1))
* **bigquery:** Add remote function options to routine metadata ([#6702](https://github.com/googleapis/google-cloud-go/issues/6702)) ([d9a437d](https://github.com/googleapis/google-cloud-go/commit/d9a437de75a5f5151cd000d8f9a6b7fc567d8551))


### Bug Fixes

* **bigquery/storage/managedwriter:** Address possible resource leak ([#6775](https://github.com/googleapis/google-cloud-go/issues/6775)) ([979440b](https://github.com/googleapis/google-cloud-go/commit/979440b43573c1cfd744d3f63d0c633c7319ad46))
* **bigquery:** Avoid stack overflow on query param with recursive types ([#6890](https://github.com/googleapis/google-cloud-go/issues/6890)) ([854ccfc](https://github.com/googleapis/google-cloud-go/commit/854ccfca259d747759d283fc0d0053893f3c8f8d)), refs [#6884](https://github.com/googleapis/google-cloud-go/issues/6884)
* **bigquery:** Bq connection auth scopes ([#6752](https://github.com/googleapis/google-cloud-go/issues/6752)) ([8e09288](https://github.com/googleapis/google-cloud-go/commit/8e09288185f721d90288c3aa873980fc44b98613)), refs [#6744](https://github.com/googleapis/google-cloud-go/issues/6744)


### Documentation

* **bigquery/storage/managedwriter/adapt:** Typo in error string ([#6729](https://github.com/googleapis/google-cloud-go/issues/6729)) ([bb26153](https://github.com/googleapis/google-cloud-go/commit/bb26153d38475cd9784edbf241df84c368f5a166))
* **bigquery/storage/managedwriter:** Add retry info to package docs ([#6803](https://github.com/googleapis/google-cloud-go/issues/6803)) ([81e52e5](https://github.com/googleapis/google-cloud-go/commit/81e52e59dcf3c4a44108e039fb0e3a0e2ce8284f))

## [1.42.0](https://github.com/googleapis/google-cloud-go/compare/bigquery/v1.41.0...bigquery/v1.42.0) (2022-09-21)


### Features

* **bigquery/analyticshub:** Start generating apiv1 ([#6707](https://github.com/googleapis/google-cloud-go/issues/6707)) ([feb7d7d](https://github.com/googleapis/google-cloud-go/commit/feb7d7d4b85d51aef6798d87a6ba8e9c536d040c))
* **bigquery/datapolicies:** Start generating apiv1beta1 ([#6697](https://github.com/googleapis/google-cloud-go/issues/6697)) ([f5443e8](https://github.com/googleapis/google-cloud-go/commit/f5443e81ae14e6aed2befe03c0df611bf5533b1f))
* **bigquery/reservation/apiv1beta1:** add REST transport ([f7b0822](https://github.com/googleapis/google-cloud-go/commit/f7b082212b1e46ff2f4126b52d49618785c2e8ca))
* **bigquery/storage/managedwriter:** Define append retry predicate ([#6650](https://github.com/googleapis/google-cloud-go/issues/6650)) ([478b8dd](https://github.com/googleapis/google-cloud-go/commit/478b8dd4e0d722cbf02fa2e216929eb561694fe0))
* **bigquery/storage:** add proto annotation for non-ascii field mapping ([ec1a190](https://github.com/googleapis/google-cloud-go/commit/ec1a190abbc4436fcaeaa1421c7d9df624042752))
* **bigquery:** Add reference file schema option for federated formats ([#6693](https://github.com/googleapis/google-cloud-go/issues/6693)) ([3d26091](https://github.com/googleapis/google-cloud-go/commit/3d26091bb8861ccfcc8d0a1727f8bbb9014ef866))
* **bigquery:** Add support for explicit query parameter type ([#6596](https://github.com/googleapis/google-cloud-go/issues/6596)) ([d59b5b2](https://github.com/googleapis/google-cloud-go/commit/d59b5b2da7d1caa6622aec84b4004cf02fb4b066)), refs [#4704](https://github.com/googleapis/google-cloud-go/issues/4704)


### Bug Fixes

* **bigquery/connection:** integrate  gapic-generator-python-1.4.1 and enable more py_test targets ([ec1a190](https://github.com/googleapis/google-cloud-go/commit/ec1a190abbc4436fcaeaa1421c7d9df624042752))

## [1.41.0](https://github.com/googleapis/google-cloud-go/compare/bigquery/v1.40.0...bigquery/v1.41.0) (2022-09-14)


### Features

* **bigquery/storage:** add location to WriteStream and add WriteStreamView support ([6a0080a](https://github.com/googleapis/google-cloud-go/commit/6a0080ad69398c572d856886293e19c79cf0fc0e))


### Bug Fixes

* **bigquery/storage/managedwriter:** Fix incorrect error retention ([#6659](https://github.com/googleapis/google-cloud-go/issues/6659)) ([dc02bca](https://github.com/googleapis/google-cloud-go/commit/dc02bca4ac14acb4f536f078a7d8f209626340bb))
* **bigquery:** Parse timestamp query parameter with RFC3339 ([#6653](https://github.com/googleapis/google-cloud-go/issues/6653)) ([aabd2d6](https://github.com/googleapis/google-cloud-go/commit/aabd2d61c81ed598755656b4e7c3fd84dcd3b2d4))

## [1.40.0](https://github.com/googleapis/google-cloud-go/compare/bigquery/v1.39.0...bigquery/v1.40.0) (2022-09-08)


### ⚠ BREAKING CHANGES

* **bigquery/dataexchange:** update BigQuery Analytics Hub API v1beta1 client BREAKING CHANGE: refresh current dataexchange/v1beta1/* directory to include recent change in protos. Removed common directory and use local enum Category fix!: refactor references to Category message docs: improve proto documentation.

### Features

* **bigquery/dataexchange:** update BigQuery Analytics Hub API v1beta1 client BREAKING CHANGE: refresh current dataexchange/v1beta1/* directory to include recent change in protos. Removed common directory and use local enum Category fix!: refactor references to Category message docs: improve proto documentation. ([e45ad9a](https://github.com/googleapis/google-cloud-go/commit/e45ad9af568c59151decc0dacedf137653b576dd))
* **bigquery/storage/managedwriter:** Augment reconnection logic ([#6609](https://github.com/googleapis/google-cloud-go/issues/6609)) ([6b0ac0c](https://github.com/googleapis/google-cloud-go/commit/6b0ac0c400d2d5b26689176c71cc6db1db9b283f))
* **bigquery:** Add trace instrumentation support for individual rpcs ([#6493](https://github.com/googleapis/google-cloud-go/issues/6493)) ([eedc632](https://github.com/googleapis/google-cloud-go/commit/eedc6327b845850d9d6109014a5d531dfbfa7d04))
* **bigquery:** Improve error when reading null values ([#6566](https://github.com/googleapis/google-cloud-go/issues/6566)) ([e9a94c2](https://github.com/googleapis/google-cloud-go/commit/e9a94c2e52ca3d07bc15030cf411f7e1c5235d39)), refs [#2612](https://github.com/googleapis/google-cloud-go/issues/2612)


### Documentation

* **bigquery:** Add numeric and bignumeric to RowIterator docs ([#6560](https://github.com/googleapis/google-cloud-go/issues/6560)) ([bea4028](https://github.com/googleapis/google-cloud-go/commit/bea4028a5fde6e790f70b0a98c33b81b3ad4023e))


### Miscellaneous Chores

* **bigquery:** Release 1.40.0 ([#6635](https://github.com/googleapis/google-cloud-go/issues/6635)) ([628deae](https://github.com/googleapis/google-cloud-go/commit/628deae4e0e0f4f1ae7e99433eefdc8f7cc41b41))

## [1.39.0](https://github.com/googleapis/google-cloud-go/compare/bigquery/v1.38.0...bigquery/v1.39.0) (2022-08-23)


### Features

* **bigquery/storage:** allow users to set Apache Avro output format options through avro_serialization_options param in TableReadOptions message Through AvroSerializationOptions, users can set enable_display_name_attribute, which populates displayName for every avro field with the original column name Improved documentation for selected_fields, added example for clarity. ([41ab4ec](https://github.com/googleapis/google-cloud-go/commit/41ab4ec00552931b12f61a9fcb27b36a7c0b5d77))
* **bigquery:** add PreserveAsciiControlCharacters support for CSV ([#6448](https://github.com/googleapis/google-cloud-go/issues/6448)) ([b7bac2f](https://github.com/googleapis/google-cloud-go/commit/b7bac2fbf63c2a681da6fdbf5af217bf8de1455f))
* **bigquery:** add preview support for default values ([#6464](https://github.com/googleapis/google-cloud-go/issues/6464)) ([edc3be5](https://github.com/googleapis/google-cloud-go/commit/edc3be586f9e8b65c34318773f5c55e1a4ccb07b))

## [1.38.0](https://github.com/googleapis/google-cloud-go/compare/bigquery/v1.37.0...bigquery/v1.38.0) (2022-08-12)


### Features

* **bigquery/migration:** Add MySQL dialect to bigquerymigration v2 client library ([370e23e](https://github.com/googleapis/google-cloud-go/commit/370e23eaa342a7055a8d8b6f8fe9420f83afe43e))
* **bigquery/storage/managedwriter:** improve error communication ([#6360](https://github.com/googleapis/google-cloud-go/issues/6360)) ([b30d89d](https://github.com/googleapis/google-cloud-go/commit/b30d89d5bd4a8ce553a328abb4b78f8fc51b43f0))


### Bug Fixes

* **bigquery/storage/managedwriter:** propagate calloptions to append ([#6488](https://github.com/googleapis/google-cloud-go/issues/6488)) ([c65f9da](https://github.com/googleapis/google-cloud-go/commit/c65f9dab8118295e49a7b863f59cb64ace4c2d5b))

## [1.37.0](https://github.com/googleapis/google-cloud-go/compare/bigquery/v1.36.0...bigquery/v1.37.0) (2022-08-04)


### Features

* **bigquery/connection:** Add service_account_id output field to CloudSQL properties ([1d6fbcc](https://github.com/googleapis/google-cloud-go/commit/1d6fbcc6406e2063201ef5a98de560bf32f7fb73))
* **bigquery/storage/managedwriter:** refactor AppendResponse ([#6402](https://github.com/googleapis/google-cloud-go/issues/6402)) ([c07bca2](https://github.com/googleapis/google-cloud-go/commit/c07bca2d65ec9903ba0c592da11440cebe8b7d9e))
* **bigquery:** support JSON as a data type ([#5986](https://github.com/googleapis/google-cloud-go/issues/5986)) ([835fe4f](https://github.com/googleapis/google-cloud-go/commit/835fe4fe59f4a3c64c5762a530228d5369618897))


### Bug Fixes

* **bigquery:** include user_email field when requesting job information ([#6256](https://github.com/googleapis/google-cloud-go/issues/6256)) ([da42b4e](https://github.com/googleapis/google-cloud-go/commit/da42b4e05faa067b5afa0a9a479d1db72296948e))


### Documentation

* **bigquery/storage:** clarify size limitations for AppendRowsRequest chore: add preferred_min_stream_count to CreateReadSessionRequest chore: add write_stream to AppendRowsResponse ([1d6fbcc](https://github.com/googleapis/google-cloud-go/commit/1d6fbcc6406e2063201ef5a98de560bf32f7fb73))

## [1.36.0](https://github.com/googleapis/google-cloud-go/compare/bigquery/v1.35.0...bigquery/v1.36.0) (2022-07-18)


### Features

* **bigquery/migration:** Add Presto dialect to bigquerymigration v2 client library ([89a049a](https://github.com/googleapis/google-cloud-go/commit/89a049a98e1d18b922cc6ad08622161448544902))
* **bigquery/storage/managedwriter/adapt:** support packed field option ([#6312](https://github.com/googleapis/google-cloud-go/issues/6312)) ([fc3417b](https://github.com/googleapis/google-cloud-go/commit/fc3417be70cd01a0044ec934c5c6426ea833d90c))


### Bug Fixes

* **bigquery/storage/managedwriter:** improve network reconnection ([#6338](https://github.com/googleapis/google-cloud-go/issues/6338)) ([085a038](https://github.com/googleapis/google-cloud-go/commit/085a03865d20122c74e107ea43883ae33bdf25bc))

## [1.35.0](https://github.com/googleapis/google-cloud-go/compare/bigquery/v1.34.1...bigquery/v1.35.0) (2022-07-07)


### Features

* **bigquery:** start generating REST client for beta clients ([25b7775](https://github.com/googleapis/google-cloud-go/commit/25b77757c1e6f372e03bf99ab7461264bba48d26))


### Bug Fixes

* **bigquery/storage/managedstorage:** improve internal locking ([#6304](https://github.com/googleapis/google-cloud-go/issues/6304)) ([a2925ce](https://github.com/googleapis/google-cloud-go/commit/a2925ce2f96c538d3994e2d0cef49fbcdd727217))
* **bigquery/storage/managedwriter/adapt:** schema->protodescriptor ([#6267](https://github.com/googleapis/google-cloud-go/issues/6267)) ([a017230](https://github.com/googleapis/google-cloud-go/commit/a01723055cb7604047c4fddd7d00213d800e4122)), refs [#6258](https://github.com/googleapis/google-cloud-go/issues/6258)
* **bigquery/storage:** Modify client lib retry policy for CreateWriteStream with longer backoff, more error code and longer overall time ([199b725](https://github.com/googleapis/google-cloud-go/commit/199b7250f474b1a6f53dcf0aac0c2966f4987b68))

## [1.34.1](https://github.com/googleapis/google-cloud-go/compare/bigquery/v1.34.0...bigquery/v1.34.1) (2022-06-23)


### Bug Fixes

* **bigquery:** release 1.34.1 ([#6251](https://github.com/googleapis/google-cloud-go/issues/6251)) ([c742b0e](https://github.com/googleapis/google-cloud-go/commit/c742b0ee644246162acaa964fc2a65eef392846f))

## [1.34.0](https://github.com/googleapis/google-cloud-go/compare/bigquery/v1.33.0...bigquery/v1.34.0) (2022-06-23)


### Features

* **bigquery/storage:** add fields to eventually contain row level errors ([5fe3b1d](https://github.com/googleapis/google-cloud-go/commit/5fe3b1d946db991aebdfd279f6f3b06b8baec205))


### Bug Fixes

* **bigquery/storage/managedwriter:** address locking and schema updates ([#6243](https://github.com/googleapis/google-cloud-go/issues/6243)) ([fe264a5](https://github.com/googleapis/google-cloud-go/commit/fe264a5ccfe5b13f6b7215d66a04282f5e38457f))

## [1.33.0](https://github.com/googleapis/google-cloud-go/compare/bigquery/v1.32.0...bigquery/v1.33.0) (2022-06-16)


### Features

* **bigquery/migration:** Add SQL Server dialect to bigquerymigration v2 client library ([90489b1](https://github.com/googleapis/google-cloud-go/commit/90489b10fd7da4cfafe326e00d1f4d81570147f7))
* **bigquery/storage/managedwriter/adapt:** support proto3 presence ([#6021](https://github.com/googleapis/google-cloud-go/issues/6021)) ([2984600](https://github.com/googleapis/google-cloud-go/commit/29846003d4bebb1a07ac11cac462ae65cb0bc59c))
* **bigquery/storage/managedwriter:** improve proto3 normalization ([#6082](https://github.com/googleapis/google-cloud-go/issues/6082)) ([6a742ff](https://github.com/googleapis/google-cloud-go/commit/6a742ffb16da991a0bf2df37218dc40af5ee0011))
* **bigquery:** add support for dataset tags ([#6114](https://github.com/googleapis/google-cloud-go/issues/6114)) ([1f35044](https://github.com/googleapis/google-cloud-go/commit/1f35044894c7395326d7a5b787a2406d5cd92cb1))
* **bigquery:** support partial projection of table metadata ([#6186](https://github.com/googleapis/google-cloud-go/issues/6186)) ([507a2be](https://github.com/googleapis/google-cloud-go/commit/507a2be8e4fda152d517dcb972be6353a6da2914))


### Bug Fixes

* **bigquery/dataexchange:** Include common protos in google-cloud-bigquery-data_exchange-v1beta1 ([6ef576e](https://github.com/googleapis/google-cloud-go/commit/6ef576e2d821d079e7b940cd5d49fe3ca64a7ba2))

## [1.32.0](https://github.com/googleapis/google-cloud-go/compare/bigquery/v1.31.0...bigquery/v1.32.0) (2022-05-06)


### Features

* **bigquery:** add interval support ([#5907](https://github.com/googleapis/google-cloud-go/issues/5907)) ([9e979c9](https://github.com/googleapis/google-cloud-go/commit/9e979c9718df1de440d440e4c3e20bb3cb8c5aa1))
* **bigquery:** expose connections and schema autodetect modifier ([#5739](https://github.com/googleapis/google-cloud-go/issues/5739)) ([c72e34f](https://github.com/googleapis/google-cloud-go/commit/c72e34fd79990eedaa56ed9e5121ab1a7fc4e2da))

## [1.31.0](https://github.com/googleapis/google-cloud-go/compare/bigquery/v1.30.2...bigquery/v1.31.0) (2022-04-12)


### Features

* **bigquery/storage:** Deprecate format specific `row_count` field in Read API ([57896d1](https://github.com/googleapis/google-cloud-go/commit/57896d1491c04fa53d3f3e2344ef10c3d91c4b65))
* **bigquery:** enhance SchemaFromJSON ([#5877](https://github.com/googleapis/google-cloud-go/issues/5877)) ([16289f0](https://github.com/googleapis/google-cloud-go/commit/16289f086ae15ea18c70d387b542796e099d4a09))
* **bigquery:** support table cloning ([#5672](https://github.com/googleapis/google-cloud-go/issues/5672)) ([74c120a](https://github.com/googleapis/google-cloud-go/commit/74c120a81d2181d9809e5d3c0462bd859297d073))

### [1.30.2](https://github.com/googleapis/google-cloud-go/compare/bigquery/v1.30.1...bigquery/v1.30.2) (2022-03-30)


### Bug Fixes

* **bigquery/storage/managedwriter/adapt:** fix enum append ([#5819](https://github.com/googleapis/google-cloud-go/issues/5819)) ([9eeaf0f](https://github.com/googleapis/google-cloud-go/commit/9eeaf0fe9de6e9583a6994e49f95ad524bc9e68e))

### [1.30.1](https://github.com/googleapis/google-cloud-go/compare/bigquery/v1.30.0...bigquery/v1.30.1) (2022-03-30)


### Bug Fixes

* **bigquery/storage/managedwriter:** correct enum processing in NormalizeDescriptor ([#5811](https://github.com/googleapis/google-cloud-go/issues/5811)) ([52cf48e](https://github.com/googleapis/google-cloud-go/commit/52cf48edff487352c2755de86e2ea069b1b29617))
* **bigquery:** improve retry for table create ([#5807](https://github.com/googleapis/google-cloud-go/issues/5807)) ([f27d1dc](https://github.com/googleapis/google-cloud-go/commit/f27d1dc43acbd437f517c05d65c992644f3f3111))

## [1.30.0](https://github.com/googleapis/google-cloud-go/compare/bigquery/v1.29.0...bigquery/v1.30.0) (2022-03-16)


### Features

* **bigquery:** support authorized datasets ([#5666](https://github.com/googleapis/google-cloud-go/issues/5666)) ([859048e](https://github.com/googleapis/google-cloud-go/commit/859048e491dd840c9aea218fa670ed2fb46d78e2))


### Bug Fixes

* **bigquery:** Query.Read fails with dry-run queries ([#5753](https://github.com/googleapis/google-cloud-go/issues/5753)) ([e279584](https://github.com/googleapis/google-cloud-go/commit/e279584727e2a496b3d566ed6f6683715a594a6d))

## [1.29.0](https://github.com/googleapis/google-cloud-go/compare/bigquery/v1.28.0...bigquery/v1.29.0) (2022-03-02)


### Features

* **bigquery/storage/managedwriter/adapt:** handle oneof normalization ([#5670](https://github.com/googleapis/google-cloud-go/issues/5670)) ([c7f54d8](https://github.com/googleapis/google-cloud-go/commit/c7f54d81baa34ce0f31bbe0af1cb03c2598e5e74))
* **bigquery/storage/managedwriter:** minor ease-of-use improvements ([#5660](https://github.com/googleapis/google-cloud-go/issues/5660)) ([d253c24](https://github.com/googleapis/google-cloud-go/commit/d253c24fd61f181971056ba00749efd69b3ae691))
* **bigquery/storage:** add trace_id for Read API ([080adb0](https://github.com/googleapis/google-cloud-go/commit/080adb0b855289ddbd86ac9e5e6eb236673f6884))
* **bigquery:** add job timeout support ([#5707](https://github.com/googleapis/google-cloud-go/issues/5707)) ([868363c](https://github.com/googleapis/google-cloud-go/commit/868363cbc68c655d4c1f8959280cf1acba5073a7))
* **bigquery:** set versionClient to module version ([55f0d92](https://github.com/googleapis/google-cloud-go/commit/55f0d92bf112f14b024b4ab0076c9875a17423c9))


### Bug Fixes

* **bigquery/storage:** remove bigquery.readonly auth scope ([5af548b](https://github.com/googleapis/google-cloud-go/commit/5af548bee4ffde279727b2e1ad9b072925106a74))

## [1.28.0](https://github.com/googleapis/google-cloud-go/compare/bigquery/v1.27.0...bigquery/v1.28.0) (2022-02-14)


### Features

* **bigquery/datatransfer:** add owner email to TransferConfig message feat: allow customer to enroll a datasource programmatically docs: improvements to various message and field descriptions ([f560b1e](https://github.com/googleapis/google-cloud-go/commit/f560b1ed0263956ef84fbf2fbf34bdc66dbc0a88))
* **bigquery:** add better version metadata to calls ([d1ad921](https://github.com/googleapis/google-cloud-go/commit/d1ad921d0322e7ce728ca9d255a3cf0437d26add))


### Bug Fixes

* **bigquery/storage/managedwriter:** address possible panic due to flow ([#5436](https://github.com/googleapis/google-cloud-go/issues/5436)) ([50c6e38](https://github.com/googleapis/google-cloud-go/commit/50c6e38c2798b3d4f2a9560239753ecd04502273))
* **bigquery/storage/managedwriter:** append improvements ([#5465](https://github.com/googleapis/google-cloud-go/issues/5465)) ([aa167bd](https://github.com/googleapis/google-cloud-go/commit/aa167bd5e57facb0f0d6834ab65805956e4ef08c))

## [1.27.0](https://www.github.com/googleapis/google-cloud-go/compare/bigquery/v1.26.0...bigquery/v1.27.0) (2022-01-24)


### Features

* **bigquery:** augment retry predicate ([#5387](https://www.github.com/googleapis/google-cloud-go/issues/5387)) ([f9608d4](https://www.github.com/googleapis/google-cloud-go/commit/f9608d4622c56362b2ed0a5845b8fe27f81995aa))
* **bigquery:** support null marker for csv in external data config ([#5287](https://www.github.com/googleapis/google-cloud-go/issues/5287)) ([132904a](https://www.github.com/googleapis/google-cloud-go/commit/132904a061809ba7117c51e8a8000f1adac34e48))

## [1.26.0](https://www.github.com/googleapis/google-cloud-go/compare/bigquery/v1.25.0...bigquery/v1.26.0) (2022-01-04)


### Features

* **bigquery/reservation:** increase the logical timeout (retry deadline) to 5 minutes ([5444809](https://www.github.com/googleapis/google-cloud-go/commit/5444809e0b7cf9f5416645ea2df6fec96f8b9023))
* **bigquery/storage/managedwriter:** support schema change notification ([#5253](https://www.github.com/googleapis/google-cloud-go/issues/5253)) ([70e40db](https://www.github.com/googleapis/google-cloud-go/commit/70e40db88bc016f228a425da1e278fc76dbf2e36))
* **bigquery/storage:** add write_mode support for BigQuery Storage Write API v1 ([615b42b](https://www.github.com/googleapis/google-cloud-go/commit/615b42bbb549b6fd3e8b1ba751bc109f79a5575b))

## [1.25.0](https://www.github.com/googleapis/google-cloud-go/compare/bigquery/v1.24.0...bigquery/v1.25.0) (2021-12-02)


### ⚠ BREAKING CHANGES

* **bigquery/storage/managedwriter:** changes function signatures to add variadic call options

### Features

* **bigquery/storage/managedwriter:** extend managedstream to support call options ([#5078](https://www.github.com/googleapis/google-cloud-go/issues/5078)) ([fbc2717](https://www.github.com/googleapis/google-cloud-go/commit/fbc2717ec84b1c5557873efaa732c047da66c1e6))
* **bigquery/storage/managedwriter:** improve method parity in managedwriter ([#5007](https://www.github.com/googleapis/google-cloud-go/issues/5007)) ([a2af4de](https://www.github.com/googleapis/google-cloud-go/commit/a2af4de215a42848368ec3081263d34782032caa))
* **bigquery/storage/managedwriter:** support variadic appends ([#5102](https://www.github.com/googleapis/google-cloud-go/issues/5102)) ([014b314](https://www.github.com/googleapis/google-cloud-go/commit/014b314b2db70147a26120a1d54a6bc7142d5665))
* **bigquery:** add BI Engine information to query statistics ([#5081](https://www.github.com/googleapis/google-cloud-go/issues/5081)) ([b78c89b](https://www.github.com/googleapis/google-cloud-go/commit/b78c89b18a81ce155441554cb5455600168eb8fd))
* **bigquery:** add support for AvroOptions in external data config ([#4945](https://www.github.com/googleapis/google-cloud-go/issues/4945)) ([8844e40](https://www.github.com/googleapis/google-cloud-go/commit/8844e40b7c2a7347e174587ea2cf438a6da9e16f))
* **bigquery:** allow construction of jobs from other projects ([#5048](https://www.github.com/googleapis/google-cloud-go/issues/5048)) ([6d07eca](https://www.github.com/googleapis/google-cloud-go/commit/6d07eca680362807f6dd870ba9df8c26256601ab))
* **bigquery:** expose identifiers using a variety of formats ([#5017](https://www.github.com/googleapis/google-cloud-go/issues/5017)) ([c9cd984](https://www.github.com/googleapis/google-cloud-go/commit/c9cd9846b6707d236648d33d44434e64eced9cdd))


### Bug Fixes

* **bigquery/migration:** correct python namespace for migration API Committer: [@shollyman](https://www.github.com/shollyman) ([8c5c6cf](https://www.github.com/googleapis/google-cloud-go/commit/8c5c6cf9df046b67998a8608d05595bd9e34feb0))
* **bigquery/storage/managedwriter:** correctly copy request ([#5122](https://www.github.com/googleapis/google-cloud-go/issues/5122)) ([cd43a5c](https://www.github.com/googleapis/google-cloud-go/commit/cd43a5cde5e4e388266f3773f206ead90d666261))
* **bigquery:** address one other callsite for the job construction feature ([#5059](https://www.github.com/googleapis/google-cloud-go/issues/5059)) ([98779eb](https://www.github.com/googleapis/google-cloud-go/commit/98779eba0f1f95b195aa6194210208767c169f5e))


### Miscellaneous Chores

* **bigquery:** release 1.25.0 ([#5128](https://www.github.com/googleapis/google-cloud-go/issues/5128)) ([f58a9f7](https://www.github.com/googleapis/google-cloud-go/commit/f58a9f7b88e2ce6101cb4bd3c85c267a688a1a1d))
* **bigquery:** release 1.25.0 ([#5177](https://www.github.com/googleapis/google-cloud-go/issues/5177)) ([359f5b1](https://www.github.com/googleapis/google-cloud-go/commit/359f5b1ca118ff6f92603da083eb943b672ed779))

## [1.24.0](https://www.github.com/googleapis/google-cloud-go/compare/bigquery/v1.23.0...bigquery/v1.24.0) (2021-09-27)


### Features

* **bigquery/migration:** Add PAUSED state to Subtask and add task details protos ([bddab08](https://www.github.com/googleapis/google-cloud-go/commit/bddab08dfd0b9a0a79b113a46a0dd84dba1f3d3b))


### Bug Fixes

* **bigquery/storage:** add missing read api retry setting on SplitReadStream ([797a9bd](https://www.github.com/googleapis/google-cloud-go/commit/797a9bdcb68c0c3ff7eef04cd3a3a0747937975b))

## [1.23.0](https://www.github.com/googleapis/google-cloud-go/compare/bigquery/v1.22.0...bigquery/v1.23.0) (2021-09-23)


### Features

* **bigquery/reservation:**
  * Deprecated SearchAssignments in favor of SearchAllAssignments
  * feat: Reservation objects now contain a creation time and an update time
  * feat: Added commitment_start_time to capacity commitments
  * feat: Force deleting capacity commitments is allowed while reservations with active assignments exist
  * feat: ML_EXTERNAL job type is supported
  * feat: Optional id can be passed into CreateCapacityCommitment and CreateAssignment
  * docs: Clarified docs for None assignments
  * fix!: Fixed pattern for BiReservation object BREAKING_CHANGE: Changed from `bireservation` to `biReservation`
  * ([d9ce9d0](https://www.github.com/googleapis/google-cloud-go/commit/d9ce9d0ee64f59c4e07ce4752bfd721051a95ac7))
* **bigquery/storage/managedwriter:** BREAKING CHANGE: changeAppendRows behavior ([#4729](https://github.com/googleapis/google-cloud-go/pull/4729))
* **bigquery/storage:** add BigQuery Storage Write API v1 ([e52c204](https://www.github.com/googleapis/google-cloud-go/commit/e52c2042a2b7cdd7dd799a561421f32fecc5d1d2))
* **bigquery/storage:** migrate managedwriter to v1 write from v1beta2 ([#4788](https://github.com/googleapis/google-cloud-go/pull/4788))
* **bigquery:** add session and connection support ([#4754](https://www.github.com/googleapis/google-cloud-go/issues/4754)) ([e846dfd](https://www.github.com/googleapis/google-cloud-go/commit/e846dfdefbba88320088667525e5fdd966c80c4b))
* **bigquery:** expose the query source of a rowiterator via SourceJob() ([#4748](https://github.com/googleapis/google-cloud-go/pull/4748))

## [1.22.0](https://www.github.com/googleapis/google-cloud-go/compare/bigquery/v1.21.0...bigquery/v1.22.0) (2021-08-30)


### Features

* **bigquery/storage/managedwriter/adapt:** add NormalizeDescriptor ([#4681](https://www.github.com/googleapis/google-cloud-go/issues/4681)) ([c54aa74](https://www.github.com/googleapis/google-cloud-go/commit/c54aa74f7a0574cbbe3f65dc90b96cf5a0b1aa88))
* **bigquery/storage/managedwriter:** more metrics instrumentation ([#4690](https://www.github.com/googleapis/google-cloud-go/issues/4690)) ([9505384](https://www.github.com/googleapis/google-cloud-go/commit/9505384b2c771d7d0c95f7786744bdf76174c706))

## [1.21.0](https://www.github.com/googleapis/google-cloud-go/compare/bigquery/v1.20.1...bigquery/v1.21.0) (2021-08-16)


### Features

* **bigquery/storage/managedwriter:** add project autodetection ([#4605](https://www.github.com/googleapis/google-cloud-go/issues/4605)) ([d8cc9be](https://www.github.com/googleapis/google-cloud-go/commit/d8cc9be6f0314f585f708638834abfc209799724))
* **bigquery/storage/managedwriter:** improve protobuf support ([#4589](https://www.github.com/googleapis/google-cloud-go/issues/4589)) ([a455082](https://www.github.com/googleapis/google-cloud-go/commit/a45508272a730e0ad81021695d2d8564e7c81631))
* **bigquery/storage/managedwriter:** more instrumentation support ([#4601](https://www.github.com/googleapis/google-cloud-go/issues/4601)) ([ff488c8](https://www.github.com/googleapis/google-cloud-go/commit/ff488c86b9c1a1f02397bb579905fa049e59ac05))
* **bigquery:** switch to centralized project autodetect logic ([#4625](https://www.github.com/googleapis/google-cloud-go/issues/4625)) ([18ff070](https://www.github.com/googleapis/google-cloud-go/commit/18ff070b8baa3ed7d324ca9ea00dcd66d7742340))


### Bug Fixes

* **bigquery/storage/managedwriter:** support non-default regions ([#4566](https://www.github.com/googleapis/google-cloud-go/issues/4566)) ([68418f9](https://www.github.com/googleapis/google-cloud-go/commit/68418f9e340def179eb5556aea433c0d07000b79))

### [1.20.1](https://www.github.com/googleapis/google-cloud-go/compare/bigquery/v1.20.0...bigquery/v1.20.1) (2021-08-06)


### Bug Fixes

* **bigquery/storage/managedwriter:** fix flowcontroller double-release ([#4555](https://www.github.com/googleapis/google-cloud-go/issues/4555)) ([67facd9](https://www.github.com/googleapis/google-cloud-go/commit/67facd9697e931e193f3cd8e188f1dd819ba31eb))

## [1.20.0](https://www.github.com/googleapis/google-cloud-go/compare/bigquery/v1.19.0...bigquery/v1.20.0) (2021-07-30)


### Features

* **bigquery/connection:** add cloud spanner connection support ([458f15b](https://www.github.com/googleapis/google-cloud-go/commit/458f15bb6f1193ce83dbfc7a82c3f2a672f52c06))
* **bigquery/storage/managedwriter/adapt:** add schema -> proto support ([#4375](https://www.github.com/googleapis/google-cloud-go/issues/4375)) ([4ff6243](https://www.github.com/googleapis/google-cloud-go/commit/4ff62433f58c1c92976a66e890b7d5394198f77b))
* **bigquery/storage/managedwriter:** add append stream plumbing ([#4452](https://www.github.com/googleapis/google-cloud-go/issues/4452)) ([b085384](https://www.github.com/googleapis/google-cloud-go/commit/b0853846a34a32ca45deb92a3cc6ab843473acd8))
* **bigquery/storage/managedwriter:** add base client ([#4422](https://www.github.com/googleapis/google-cloud-go/issues/4422)) ([4f7193b](https://www.github.com/googleapis/google-cloud-go/commit/4f7193b74f4b1954cf7b664d61b5cc9805337e84))
* **bigquery/storage/managedwriter:** add flow controller ([#4404](https://www.github.com/googleapis/google-cloud-go/issues/4404)) ([9dc78e0](https://www.github.com/googleapis/google-cloud-go/commit/9dc78e073b5f69037c6328460554c4354fcee11f))
* **bigquery/storage/managedwriter:** add opencensus instrumentation ([#4512](https://www.github.com/googleapis/google-cloud-go/issues/4512)) ([73b6f5e](https://www.github.com/googleapis/google-cloud-go/commit/73b6f5e012d0b89d36850cb986fd7e288bf1e3c5))
* **bigquery/storage/managedwriter:** add state tracking ([#4407](https://www.github.com/googleapis/google-cloud-go/issues/4407)) ([4638e17](https://www.github.com/googleapis/google-cloud-go/commit/4638e17dacd1fa76f9976f44974c4037fe4358dc))
* **bigquery/storage/managedwriter:** naming and doc improvements ([#4508](https://www.github.com/googleapis/google-cloud-go/issues/4508)) ([663c899](https://www.github.com/googleapis/google-cloud-go/commit/663c899c3b8aa751527d24f541d964f2ba91a233))
* **bigquery/storage/managedwriter:** wire in flow controller ([#4501](https://www.github.com/googleapis/google-cloud-go/issues/4501)) ([40571fa](https://www.github.com/googleapis/google-cloud-go/commit/40571fa0e3b5ab326fd592a6907061c2304893aa))
* **bigquery:** add more dml statistics to query statistics ([#4405](https://www.github.com/googleapis/google-cloud-go/issues/4405)) ([99d5728](https://www.github.com/googleapis/google-cloud-go/commit/99d57282f6668de91390ad29a888a89209689f39))
* **bigquery:** support decimalTargetType prioritization ([#4343](https://www.github.com/googleapis/google-cloud-go/issues/4343)) ([95a27f7](https://www.github.com/googleapis/google-cloud-go/commit/95a27f711a1c7dfdaa16ae5d3c52644769b6fc39))
* **bigquery:** support multistatement transaction statistics in jobs ([#4485](https://www.github.com/googleapis/google-cloud-go/issues/4485)) ([4565eb7](https://www.github.com/googleapis/google-cloud-go/commit/4565eb7fe730eade294fb3baa85bd255df008bfa))


### Bug Fixes

* **bigquery/storage/managedwriter:** fix double-close error, add tests ([#4502](https://www.github.com/googleapis/google-cloud-go/issues/4502)) ([c6cf659](https://www.github.com/googleapis/google-cloud-go/commit/c6cf6590a41368885b7399c993c47dc965862558))

## [1.19.0](https://www.github.com/googleapis/google-cloud-go/compare/bigquery/v1.18.0...bigquery/v1.19.0) (2021-06-29)


### Features

* **bigquery/storage:** Add ZSTD compression as an option for Arrow. ([770db30](https://www.github.com/googleapis/google-cloud-go/commit/770db3083270d485d265362fe5a4b2a1b23619ff))
* **bigquery/storage:** remove alpha client ([#4100](https://www.github.com/googleapis/google-cloud-go/issues/4100)) ([a2d137d](https://www.github.com/googleapis/google-cloud-go/commit/a2d137d233e7a401976fbe1fd8ff81145dda515d)), refs [#4098](https://www.github.com/googleapis/google-cloud-go/issues/4098)
* **bigquery:** add support for parameterized types ([#4103](https://www.github.com/googleapis/google-cloud-go/issues/4103)) ([a2330e4](https://www.github.com/googleapis/google-cloud-go/commit/a2330e4d66c0a1832fb3b9e23a33c006c9345c28))
* **bigquery:** add support for snapshot/restore ([#4112](https://www.github.com/googleapis/google-cloud-go/issues/4112)) ([4c12b42](https://www.github.com/googleapis/google-cloud-go/commit/4c12b424eec06c7d87244eaa922995bbe6e46e7e))
* **bigquery:** add support for user defined TVF ([#4043](https://www.github.com/googleapis/google-cloud-go/issues/4043)) ([37607b4](https://www.github.com/googleapis/google-cloud-go/commit/37607b4afbc4c42baa4a931a9a86cddcc6d885ca))
* **bigquery:** enable project autodetection, expose project ids further ([#4312](https://www.github.com/googleapis/google-cloud-go/issues/4312)) ([267787e](https://www.github.com/googleapis/google-cloud-go/commit/267787eb245d9307cf78304c1ce34bdfb2aaf5ab))
* **bigquery:** support job deletion ([#3935](https://www.github.com/googleapis/google-cloud-go/issues/3935)) ([363ba03](https://www.github.com/googleapis/google-cloud-go/commit/363ba03e1c3c813749a65ff3c050877ce4f60016))
* **bigquery:** support nullable params and geography params ([#4225](https://www.github.com/googleapis/google-cloud-go/issues/4225)) ([43755d3](https://www.github.com/googleapis/google-cloud-go/commit/43755d38b5d928222127cc6be26183d6bfbb1cb4))


### Bug Fixes

* **bigquery:** minor rename to feature that's not yet in a release ([#4320](https://www.github.com/googleapis/google-cloud-go/issues/4320)) ([ef8d138](https://www.github.com/googleapis/google-cloud-go/commit/ef8d1386149cff28ae6258ab167789bae6af6407))
* **bigquery:** update streaming insert error test ([#4321](https://www.github.com/googleapis/google-cloud-go/issues/4321)) ([12f3042](https://www.github.com/googleapis/google-cloud-go/commit/12f3042716d51fb0d7a23071d00a20f9751bac91))

## [1.18.0](https://www.github.com/googleapis/google-cloud-go/compare/bigquery/v1.17.0...bigquery/v1.18.0) (2021-05-06)


### Features

* **bigquery/storage:** new JSON type through BigQuery Write ([9029071](https://www.github.com/googleapis/google-cloud-go/commit/90290710158cf63de918c2d790df48f55a23adc5))
* **bigquery:** augment retry predicate to support additional errors ([#4046](https://www.github.com/googleapis/google-cloud-go/issues/4046)) ([d4af6f7](https://www.github.com/googleapis/google-cloud-go/commit/d4af6f7707b3c5ee12cde53c7485a9b743034119))
* **bigquery:** expose ParquetOptions for loads and external tables ([#4016](https://www.github.com/googleapis/google-cloud-go/issues/4016)) ([f9c4ccb](https://www.github.com/googleapis/google-cloud-go/commit/f9c4ccb6efb271c421edf3f67d5249b1cfb0ecb2))
* **bigquery:** support mutable clustering configuration ([#3950](https://www.github.com/googleapis/google-cloud-go/issues/3950)) ([0ab30da](https://www.github.com/googleapis/google-cloud-go/commit/0ab30dadc43ae85354dc12a4130ecfcc56273882))

## [1.17.0](https://www.github.com/googleapis/google-cloud-go/compare/bigquery/v1.15.0...bigquery/v1.17.0) (2021-04-08)


### Features

* **bigquery/storage:** add a Arrow compression options (Only LZ4 for now). feat: Return schema on first ReadRowsResponse. doc: clarify limit on filter string. ([2b02a03](https://www.github.com/googleapis/google-cloud-go/commit/2b02a03ff9f78884da5a8e7b64a336014c61bde7))
* **bigquery/storage:** deprecate bigquery storage v1alpha2 API ([9cc6d2c](https://www.github.com/googleapis/google-cloud-go/commit/9cc6d2cce96235b0a144c1c6b48eff496f9e5fa7))
* **bigquery/storage:** updates for v1beta2 storage API - Updated comments on BatchCommitWriteStreams - Added new support Bigquery types BIGNUMERIC and INTERVAL to TableSchema - Added read rows schema in ReadRowsResponse - Misc comment updates ([48b4e59](https://www.github.com/googleapis/google-cloud-go/commit/48b4e596206cef879194d2888186d603a6f51292))
* **bigquery:** export HivePartitioningOptions in load job configurations ([#3877](https://www.github.com/googleapis/google-cloud-go/issues/3877)) ([7c759be](https://www.github.com/googleapis/google-cloud-go/commit/7c759be074ce1f6b8ccce88c86dbe49bd38fd6b5))
* **bigquery:** support type alias names for numeric/bignumeric schemas. ([#3760](https://www.github.com/googleapis/google-cloud-go/issues/3760)) ([2ee6bf4](https://www.github.com/googleapis/google-cloud-go/commit/2ee6bf451524fc1f9735634320a55ca0b07d3d8b))

## v1.16.0

- Updates to various dependencies.

## [1.15.0](https://www.github.com/googleapis/google-cloud-go/compare/bigquery/v1.14.0...v1.15.0) (2021-01-14)


### Features

* **bigquery:** add reservation usage stats to query statistics ([#3403](https://www.github.com/googleapis/google-cloud-go/issues/3403)) ([112bcde](https://www.github.com/googleapis/google-cloud-go/commit/112bcdeb7cee1b44f337d3e5398a0d0820e93162))
* **bigquery:** add support for allowing Javascript UDFs to indicate determinism ([#3534](https://www.github.com/googleapis/google-cloud-go/issues/3534)) ([2f417a3](https://www.github.com/googleapis/google-cloud-go/commit/2f417a39d93402fbb1e5e3001645019782d7d656)), refs [#3533](https://www.github.com/googleapis/google-cloud-go/issues/3533)


### Bug Fixes

* **bigquery:** address possible panic due to offset checking in handleInsertErrors  ([#3524](https://www.github.com/googleapis/google-cloud-go/issues/3524)) ([5288511](https://www.github.com/googleapis/google-cloud-go/commit/52885115af3e95cdfd1ec784837fb1df7fe01446)), refs [#3519](https://www.github.com/googleapis/google-cloud-go/issues/3519)

## [1.14.0](https://www.github.com/googleapis/google-cloud-go/compare/bigquery/v1.13.0...v1.14.0) (2020-12-04)


### Features

* **bigquery:** add support for bignumeric ([#2779](https://www.github.com/googleapis/google-cloud-go/issues/2779)) ([ea3cde5](https://www.github.com/googleapis/google-cloud-go/commit/ea3cde55ad3d8d843bce8d023747cf69552850b5))
* **bigquery:** expose hive partitioning options ([#3240](https://www.github.com/googleapis/google-cloud-go/issues/3240)) ([fa77efa](https://www.github.com/googleapis/google-cloud-go/commit/fa77efa1a1880ff89307d54cc7e9e8c09430e4e2))

## v1.13.0

* Support retries for specific http2 transport race.
* Remove unused datasource client from bigquery/datatransfer.
* Adds support for authorized User Defined Functions (UDFs).
* Documentation improvements.
* Various updates to autogenerated clients.


## v1.12.0

* Adds additional retry support for table deletion.
* Various updates to autogenerated clients.

## v1.11.2

* Addresses issue with consuming query results using an iterator.Pager

## v1.11.1

* Addresses issue with optimized query path changes, released
  in v1.11.0

## v1.11.0

* Add support for optimized query path.
* Documentation improvements.
* Fix issue related to the ReturnType of a bigquery Routine.
* Various updates to autogenerated clients.

## v1.10.0

* Support for Infinity/-Infinity/NaN values in NullFloat64.
* Updates to RowIterator to address issues related to retrieving query
  results without explicit destination table references.
* Various updates to autogenerated clients.

## v1.9.0

* SchemaFromJSON will now accept alias type names (e.g. INT64 vs INTEGER, STRUCT vs RECORD).
* Support for IAM on table resources.
* Various updates to autogenerated clients.

## v1.8.0

* Add support for hourly time partitioning.
* Various updates to autogenerated clients.

## v1.7.0

* Add support for extracting BQML models to cloud storage.
* Add support for specifying projected fields when ingesting
  datastore backups.
* Fix issue related to defining a range partitioning range
  using default values.
* Add bigquery/reservation/v1 API.
* Various updates to autogenerated clients.

## v1.6.0

* Add support for materialized views.
* Add support for policy tags (column ACLs).
* Add bigquery/connection/v1beta1 API.
* Documentation improvements.
* Various updates to autogenerated clients.

## v1.5.0

* Add v1 endpoint for bigquerystorage API.
* Improved error message in bigquery.PutMultiError.
* Various updates to autogenerated clients.

## v1.4.0

* Add v1beta2, v1alpha2 endpoints for bigquerystorage API.

* Location is now reported as part of TableMetadata.

## v1.3.0

* Add Description field for Routine entities.

* Add support for iamMember entities on dataset ACLs.

* Address issue when constructing a Pager from a RowIterator
  that referenced a result with zero result rows.

* Add support for integer range partitioning, which affects
  table creation directly and via query/load jobs.

* Add opt-out support for streaming inserts via experimental
  `NoDedupeID` sentinel.

## v1.2.0

* Adds support for scripting feature, which includes script statistics
  and the ability to list jobs run as part of a script query.

* Updates default endpoint for BigQuery from www.googleapis.com
  to bigquery.googleapis.com.

## v1.1.0

* Added support for specifying default `EncryptionConfig` settings on the
  dataset.

* Added support for `EncyptionConfig` as part of an ML model.

* Added `Relax()` to make all fields within a `Schema` nullable.

* Added a `UseAvroLogicalTypes` option when defining an avro extract job.

## v1.0.1

This patch release is a small fix to the go.mod to point to the post-carve out
cloud.google.com/go.

## v1.0.0

This is the first tag to carve out bigquery as its own module. See:
https://github.com/golang/go/wiki/Modules#is-it-possible-to-add-a-module-to-a-multi-module-repository.
