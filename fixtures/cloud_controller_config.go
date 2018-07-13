package fixtures

const CloudControllerBaseConfigYaml = `---
local_route: 127.0.0.1
tls_port: 8182
pid_filename: /tmp/cloud_controller.pid
stacks_file: config/stacks.yml

external_protocol: http
external_domain: api2.vcap.me
temporary_disable_deployments: true
internal_service_hostname: api.internal.cf

system_domain: vcap.me
app_domains: []
system_hostnames: []

jobs:
  global:
    timeout_in_seconds: 14400

default_app_memory: 1024 #mb
default_app_disk_in_mb: 1024
maximum_app_disk_in_mb: 2048

broker_client_default_async_poll_interval_seconds: 60
broker_client_max_async_poll_duration_minutes: 10080

shared_isolation_segment_name: 'shared'

info:
  name: "vcap"
  build: "2222"
  version: 2
  support_address: "http://support.cloudfoundry.com"
  description: "Cloud Foundry sponsored by Pivotal"
  app_ssh_endpoint: "ssh.system.domain.example.com:2222"
  app_ssh_oauth_client: "ssh-proxy"

instance_file_descriptor_limit: 16384
login:
  url: 'login-url.example.com'

nginx:
  use_nginx: false
  instance_socket: "/var/vcap/sys/run/cloud_controller_ng/cloud_controller.sock"

logging:
  file: /tmp/cloud_controller.log
  level: debug2
  syslog: vcap.example

loggregator:
  router: "127.0.0.1:3456"
  internal_url: 'http://loggregator-trafficcontroller.service.cf.internal:8081'

doppler:
  url: 'wss://doppler.example.com:443'

uaa:
  url: "http://localhost:6789"
  internal_url: "http://localhost:6789"
  resource_id: "cloud_controller"
  symmetric_secret: "tokensecret"
  ca_file: "spec/fixtures/certs/uaa_ca.crt"
  client_timeout: 60

route_services_enabled: true
volume_services_enabled: true

bulk_api:
  auth_user: bulk_user
  auth_password: bulk_password

internal_api:
  auth_user: internal_user
  auth_password: internal_password

# App staging parameters
staging:
  timeout_in_seconds: 120 # secs
  minimum_staging_memory_mb: 1024
  minimum_staging_disk_mb: 4096
  minimum_staging_file_descriptor_limit: 4200
  auth:
    user: zxsfhgjg
    password: ZNVfdase9

quota_definitions: {}

default_quota_definition: default

db_encryption_key: "asdfasdfasdf"

default_health_check_timeout: 60
maximum_health_check_timeout: 180

disable_custom_buildpacks: false
broker_client_timeout_seconds: 60

cloud_controller_username_lookup_client_name: 'username_lookup_client_name'
cloud_controller_username_lookup_client_secret: 'username_lookup_secret'

cc_service_key_client_name: 'cc_service_key_client'
cc_service_key_client_secret: 'cc-service-key-client-super-s3cre7'

allow_app_ssh_access: true
default_app_ssh_access: true

renderer:
  max_results_per_page: 100
  default_results_per_page: 50
  max_inline_relations_depth: 3

install_buildpacks: []

security_group_definitions: []

default_staging_security_groups: []
default_running_security_groups: []

allowed_cors_domains: []

rate_limiter:
  enabled: false
  general_limit: 2000
  unauthenticated_limit: 100
  reset_interval_in_minutes: 60

diego:
  file_server_url: http://file-server.service.cf.internal:8080
  cc_uploader_url: http://cc-uploader.service.cf.internal:9090
  use_privileged_containers_for_running: false
  use_privileged_containers_for_staging: false
  lifecycle_bundles: {}
  insecure_docker_registry_list: []
  docker_staging_stack: 'cflinuxfs2'
  bbs:
    url: https://bbs.service.cf.internal:8889
    key_file: /var/vcap/jobs/cloud_controller_ng/config/certs/bbs_client.key
    cert_file: /var/vcap/jobs/cloud_controller_ng/config/certs/bbs_client.crt
    ca_file: /var/vcap/jobs/cloud_controller_ng/config/certs/bbs_ca.crt

  pid_limit: 2048

directories:
  tmpdir: /tmp
  diagnostics: /tmp

newrelic_enabled: false

index: 0
name: api
resource_pool:
  fog_aws_storage_options: {}
  fog_connection: {}
  maximum_size: 42
  minimum_size: 1
  resource_directory_key: ''

buildpacks:
  buildpack_directory_key: cc-buildpacks
  fog_aws_storage_options: {}
  fog_connection: {}

packages:
  app_package_directory_key: "cc-packages"
  fog_aws_storage_options: {}
  fog_connection: {}
  max_package_size: 42
  max_valid_packages_stored: 42

droplets:
  droplet_directory_key: cc-droplets
  max_staged_droplets_stored: 42
  fog_aws_storage_options: {}
  fog_connection: {}

request_timeout_in_seconds: 600
skip_cert_verify: true
app_bits_upload_grace_period_in_seconds: 500
security_event_logging:
  enabled: false
  file: /tmp/cef.log

bits_service:
  enabled: false

statsd_host: "127.0.0.1"
statsd_port: 8125

credential_references:
  interpolate_service_bindings: true

db:
  log_level: 'debug'
  max_connections: 42
  pool_timeout: 60
  log_db_queries: true
  connection_validation_timeout: 3600
  database: postgres://postgres@localhost:5432/cc_test_integration_cc
`
