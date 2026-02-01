# Dependencies

## Summary

- **Production:** 84 packages
- **Development:** 21 packages
- **Total:** 105 packages

## Dependency Graph

```mermaid
graph TD
  Root[Your Application]
  Root --> aws_aws_crt_php["aws/aws-crt-php<br/>v1.2.7"]
  Root --> aws_aws_sdk_php["aws/aws-sdk-php<br/>3.369.20"]
  aws_aws_sdk_php --> aws_aws_crt_php["aws/aws-crt-php<br/>^1.2.3"]
  aws_aws_sdk_php --> guzzlehttp_guzzle["guzzlehttp/guzzle<br/>^7.4.5"]
  aws_aws_sdk_php --> guzzlehttp_promises["guzzlehttp/promises<br/>^2.0"]
  Root --> azuracast_doctrine_entity_normalizer["azuracast/doctrine-entity-normalizer<br/>3.3.0"]
  azuracast_doctrine_entity_normalizer --> doctrine_collections["doctrine/collections<br/>>1"]
  azuracast_doctrine_entity_normalizer --> doctrine_inflector["doctrine/inflector<br/>^2"]
  azuracast_doctrine_entity_normalizer --> doctrine_orm["doctrine/orm<br/>^3"]
  Root --> azuracast_nowplaying["azuracast/nowplaying<br/>dev-main"]
  azuracast_nowplaying --> guzzlehttp_guzzle["guzzlehttp/guzzle<br/>>7"]
  azuracast_nowplaying --> psr_http_client["psr/http-client<br/>*"]
  azuracast_nowplaying --> psr_http_factory["psr/http-factory<br/>*"]
  Root --> beberlei_doctrineextensions["beberlei/doctrineextensions<br/>v1.5.0"]
  beberlei_doctrineextensions --> doctrine_orm["doctrine/orm<br/>^2.19 || ^3.0"]
  Root --> br33f_php_ga4_mp["br33f/php-ga4-mp<br/>v0.1.6"]
  br33f_php_ga4_mp --> guzzlehttp_guzzle["guzzlehttp/guzzle<br/>^6.5.5 || ^7.0.0"]
  Root --> brick_math["brick/math<br/>0.14.1"]
  Root --> carbonphp_carbon_doctrine_types["carbonphp/carbon-doctrine-types<br/>3.2.0"]
  Root --> composer_ca_bundle["composer/ca-bundle<br/>1.5.10"]
  Root --> dflydev_fig_cookies["dflydev/fig-cookies<br/>v3.2.0"]
  dflydev_fig_cookies --> psr_http_message["psr/http-message<br/>^1.0.1 || ^2"]
  Root --> doctrine_collections["doctrine/collections<br/>2.6.0"]
  doctrine_collections --> doctrine_deprecations["doctrine/deprecations<br/>^1"]
  doctrine_collections --> symfony_polyfill_php84["symfony/polyfill-php84<br/>^1.30"]
  Root --> doctrine_data_fixtures["doctrine/data-fixtures<br/>2.2.0"]
  doctrine_data_fixtures --> doctrine_persistence["doctrine/persistence<br/>^3.1 || ^4.0"]
  doctrine_data_fixtures --> psr_log["psr/log<br/>^1.1 || ^2 || ^3"]
  Root --> doctrine_dbal["doctrine/dbal<br/>4.4.1"]
  doctrine_dbal --> doctrine_deprecations["doctrine/deprecations<br/>^1.1.5"]
  doctrine_dbal --> psr_cache["psr/cache<br/>^1|^2|^3"]
  doctrine_dbal --> psr_log["psr/log<br/>^1|^2|^3"]
  Root --> doctrine_deprecations["doctrine/deprecations<br/>1.1.5"]
  Root --> doctrine_event_manager["doctrine/event-manager<br/>2.1.0"]

```

## Production Dependencies

| Package | Version |
|---------|----------|
| `ext-PDO` | `*` |
| `ext-ctype` | `*` |
| `ext-curl` | `*` |
| `ext-ffi` | `*` |
| `ext-fileinfo` | `*` |
| `ext-gd` | `*` |
| `ext-iconv` | `*` |
| `ext-intl` | `*` |
| `ext-json` | `*` |
| `ext-libxml` | `*` |
| `ext-maxminddb` | `*` |
| `ext-mbstring` | `*` |
| `ext-redis` | `*` |
| `ext-simplexml` | `*` |
| `ext-xml` | `*` |
| `ext-xmlreader` | `*` |
| `ext-xmlwriter` | `*` |
| `azuracast/doctrine-entity-normalizer` | `^3.3.0` |
| `azuracast/nowplaying` | `dev-main` |
| `beberlei/doctrineextensions` | `^1.5` |
| `br33f/php-ga4-mp` | `^0.1.5` |
| `brick/math` | `^0.14` |
| `composer/ca-bundle` | `^1.5.8` |
| `doctrine/data-fixtures` | `^2.2` |
| `doctrine/dbal` | `^4.3.4` |
| `doctrine/migrations` | `^3.9.4` |
| `doctrine/orm` | `^3.5.3` |
| `dragonmantank/cron-expression` | `^3.4` |
| `gettext/gettext` | `^5.7.3` |
| `gettext/php-scanner` | `^2.0.1` |
| `gettext/translator` | `^1.2.1` |
| `guzzlehttp/guzzle` | `^7.10` |
| `intervention/image` | `^3.11.4` |
| `james-heinrich/getid3` | `v2.0.0-beta6` |
| `lbuchs/webauthn` | `^2.2` |
| `league/csv` | `^9.27.1` |
| `league/flysystem-aws-s3-v3` | `^3.30.1` |
| `league/flysystem-sftp-v3` | `^3.30` |
| `league/mime-type-detection` | `^1.16` |
| `league/oauth2-client` | `^2.8.1` |
| `league/plates` | `^3.6` |
| `lstrojny/fxmlrpc` | `dev-master` |
| `matomo/device-detector` | `^6.4.7` |
| `mezzio/mezzio-session` | `^1.17` |
| `mezzio/mezzio-session-cache` | `^1.17` |
| `monolog/monolog` | `^3.9` |
| `myclabs/deep-copy` | `^1.13.4` |
| `nesbot/carbon` | `^3.10.3` |
| `pagerfanta/doctrine-collections-adapter` | `^4.7.2` |
| `pagerfanta/doctrine-orm-adapter` | `^4.7.2` |

*... and 34 more*

## Development Dependencies

| Package | Version |
|---------|----------|
| `codeception/codeception` | `^5.3.2` |
| `codeception/module-asserts` | `^3.2` |
| `codeception/module-cli` | `^2.0.1` |
| `codeception/module-doctrine` | `^3.2` |
| `codeception/module-phpbrowser` | `dev-master` |
| `codeception/module-rest` | `^3.4.1` |
| `filp/whoops` | `^2.18.4` |
| `maxmind-db/reader` | `^1.12.1` |
| `mockery/mockery` | `^1.6.12` |
| `nette/php-generator` | `^4.2` |
| `php-parallel-lint/php-console-highlighter` | `^1` |
| `php-parallel-lint/php-parallel-lint` | `^1.4` |
| `phpstan/phpstan` | `^2.1.31` |
| `phpstan/phpstan-doctrine` | `^2.0.10` |
| `phpunit/phpunit` | `^12.4.1` |
| `psy/psysh` | `^0.12.14` |
| `pyrech/composer-changelogs` | `^2.1` |
| `roave/security-advisories` | `dev-latest` |
| `slevomat/coding-standard` | `^8.24` |
| `squizlabs/php_codesniffer` | `^4` |
| `symfony/var-dumper` | `^8` |
