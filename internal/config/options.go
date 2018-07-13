package config

type Option func(*config)

func WithPort(port int) Option {
	return func(c *config) {
		c.ExternalPort = port
	}
}

func WithPermEnabled(enabled bool) Option {
	return func(c *config) {
		c.Perm.Enabled = enabled
	}
}

func WithPermHostname(hostname string) Option {
	return func(c *config) {
		c.Perm.Hostname = hostname
	}
}

func WithPermPort(port int) Option {
	return func(c *config) {
		c.Perm.Port = port
	}
}

func WithPermCACertPath(caPath string) Option {
	return func(c *config) {
		c.Perm.CACertPath = caPath
	}
}

func WithPermTimeoutInMilliseconds(timeout int) Option {
	return func(c *config) {
		c.Perm.TimeoutInMilliseconds = timeout
	}
}

func WithUAAURL(url string) Option {
	return func(c *config) {
		c.UAA.URL = url
	}
}

func WithUAAInternalURL(internalURL string) Option {
	return func(c *config) {
		c.UAA.InternalURL = internalURL
	}
}

type config struct {
	ExternalPort                int           `yaml:"external_port"`
	LocalRoute                  string        `yaml:"local_route"`
	TLSPort                     int           `yaml:"tls_port"`
	PidFilename                 string        `yaml:"pid_filename"`
	StacksFile                  string        `yaml:"stacks_file"`
	ExternalProtocol            string        `yaml:"external_protocol"`
	ExternalDomain              string        `yaml:"external_domain"`
	TemporaryDisableDeployments bool          `yaml:"temporary_disable_deployments"`
	InternalServiceHostname     string        `yaml:"internal_service_hostname"`
	SystemDomain                string        `yaml:"system_domain"`
	AppDomains                  []interface{} `yaml:"app_domains"`
	SystemHostnames             []interface{} `yaml:"system_hostnames"`
	Jobs                        struct {
		Global struct {
			TimeoutInSeconds int `yaml:"timeout_in_seconds"`
		} `yaml:"global"`
	} `yaml:"jobs"`
	DefaultAppMemory                            int    `yaml:"default_app_memory"`
	DefaultAppDiskInMB                          int    `yaml:"default_app_disk_in_mb"`
	MaximumAppDiskInMB                          int    `yaml:"maximum_app_disk_in_mb"`
	BrokerClientDefaultAsyncPollIntervalSeconds int    `yaml:"broker_client_default_async_poll_interval_seconds"`
	BrokerClientMaxAsyncPollDurationMinutes     int    `yaml:"broker_client_max_async_poll_duration_minutes"`
	SharedIsolationSegmentName                  string `yaml:"shared_isolation_segment_name"`
	Info                                        struct {
		Name              string `yaml:"name"`
		Build             string `yaml:"build"`
		Version           int    `yaml:"version"`
		SupportAddress    string `yaml:"support_address"`
		Description       string `yaml:"description"`
		AppSSHEndpoint    string `yaml:"app_ssh_endpoint"`
		AppSSHOauthClient string `yaml:"app_ssh_oauth_client"`
	} `yaml:"info"`
	InstanceFileDescriptorLimit int `yaml:"instance_file_descriptor_limit"`
	Login                       struct {
		URL string `yaml:"url"`
	} `yaml:"login"`
	Nginx struct {
		UseNginx       bool   `yaml:"use_nginx"`
		InstanceSocket string `yaml:"instance_socket"`
	} `yaml:"nginx"`
	Logging struct {
		File   string `yaml:"file"`
		Level  string `yaml:"level"`
		Syslog string `yaml:"syslog"`
	} `yaml:"logging"`
	Loggregator struct {
		Router      string `yaml:"router"`
		InternalURL string `yaml:"internal_url"`
	} `yaml:"loggregator"`
	Doppler struct {
		URL string `yaml:"url"`
	} `yaml:"doppler"`
	UAA struct {
		URL             string `yaml:"url"`
		InternalURL     string `yaml:"internal_url"`
		ResourceID      string `yaml:"resource_id"`
		SymmetricSecret string `yaml:"symmetric_secret"`
		CAFile          string `yaml:"ca_file"`
		ClientTimeout   int    `yaml:"client_timeout"`
	} `yaml:"uaa"`
	RouteServicesEnabled  bool `yaml:"route_services_enabled"`
	VolumeServicesEnabled bool `yaml:"volume_services_enabled"`
	BulkAPI               struct {
		AuthUser     string `yaml:"auth_user"`
		AuthPassword string `yaml:"auth_password"`
	} `yaml:"bulk_api"`
	InternalAPI struct {
		AuthUser     string `yaml:"auth_user"`
		AuthPassword string `yaml:"auth_password"`
	} `yaml:"internal_api"`
	Staging struct {
		TimeoutInSeconds                  int `yaml:"timeout_in_seconds"`
		MinimumStagingMemoryMB            int `yaml:"minimum_staging_memory_mb"`
		MinimumStagingDiskMB              int `yaml:"minimum_staging_disk_mb"`
		MinimumStagingFileDescriptorLimit int `yaml:"minimum_staging_file_descriptor_limit"`
		Auth                              struct {
			User     string `yaml:"user"`
			Password string `yaml:"password"`
		} `yaml:"auth"`
	} `yaml:"staging"`
	QuotaDefinitions struct {
	} `yaml:"quota_definitions"`
	DefaultQuotaDefinition                    string `yaml:"default_quota_definition"`
	DbEncryptionKey                           string `yaml:"db_encryption_key"`
	DefaultHealthCheckTimeout                 int    `yaml:"default_health_check_timeout"`
	MaximumHealthCheckTimeout                 int    `yaml:"maximum_health_check_timeout"`
	DisableCustomBuildpacks                   bool   `yaml:"disable_custom_buildpacks"`
	BrokerClientTimeoutSeconds                int    `yaml:"broker_client_timeout_seconds"`
	CloudControllerUsernameLookupClientName   string `yaml:"cloud_controller_username_lookup_client_name"`
	CloudControllerUsernameLookupClientSecret string `yaml:"cloud_controller_username_lookup_client_secret"`
	CcServiceKeyClientName                    string `yaml:"cc_service_key_client_name"`
	CcServiceKeyClientSecret                  string `yaml:"cc_service_key_client_secret"`
	AllowAppSSHAccess                         bool   `yaml:"allow_app_ssh_access"`
	DefaultAppSSHAccess                       bool   `yaml:"default_app_ssh_access"`
	Renderer                                  struct {
		MaxResultsPerPage       int `yaml:"max_results_per_page"`
		DefaultResultsPerPage   int `yaml:"default_results_per_page"`
		MaxInlineRelationsDepth int `yaml:"max_inline_relations_depth"`
	} `yaml:"renderer"`
	InstallBuildpacks            []interface{} `yaml:"install_buildpacks"`
	SecurityGroupDefinitions     []interface{} `yaml:"security_group_definitions"`
	DefaultStagingSecurityGroups []interface{} `yaml:"default_staging_security_groups"`
	DefaultRunningSecurityGroups []interface{} `yaml:"default_running_security_groups"`
	AllowedCorsDomains           []interface{} `yaml:"allowed_cors_domains"`
	RateLimiter                  struct {
		Enabled                bool `yaml:"enabled"`
		GeneralLimit           int  `yaml:"general_limit"`
		UnauthenticatedLimit   int  `yaml:"unauthenticated_limit"`
		ResetIntervalInMinutes int  `yaml:"reset_interval_in_minutes"`
	} `yaml:"rate_limiter"`
	Diego struct {
		FileServerURL                     string `yaml:"file_server_url"`
		CcUploaderURL                     string `yaml:"cc_uploader_url"`
		UsePrivilegedContainersForRunning bool   `yaml:"use_privileged_containers_for_running"`
		UsePrivilegedContainersForStaging bool   `yaml:"use_privileged_containers_for_staging"`
		LifecycleBundles                  struct {
		} `yaml:"lifecycle_bundles"`
		InsecureDockerRegistryList []interface{} `yaml:"insecure_docker_registry_list"`
		DockerStagingStack         string        `yaml:"docker_staging_stack"`
		BBS                        struct {
			URL      string `yaml:"url"`
			KeyFile  string `yaml:"key_file"`
			CertFile string `yaml:"cert_file"`
			CaFile   string `yaml:"ca_file"`
		} `yaml:"bbs"`
		PidLimit int `yaml:"pid_limit"`
	} `yaml:"diego"`
	Directories struct {
		Tmpdir      string `yaml:"tmpdir"`
		Diagnostics string `yaml:"diagnostics"`
	} `yaml:"directories"`
	NewRelicEnabled bool   `yaml:"newrelic_enabled"`
	Index           int    `yaml:"index"`
	Name            string `yaml:"name"`
	ResourcePool    struct {
		FogAwsStorageOptions struct {
		} `yaml:"fog_aws_storage_options"`
		FogConnection struct {
		} `yaml:"fog_connection"`
		MaximumSize          int    `yaml:"maximum_size"`
		MinimumSize          int    `yaml:"minimum_size"`
		ResourceDirectoryKey string `yaml:"resource_directory_key"`
	} `yaml:"resource_pool"`
	Buildpacks struct {
		BuildpackDirectoryKey string `yaml:"buildpack_directory_key"`
		FogAwsStorageOptions  struct {
		} `yaml:"fog_aws_storage_options"`
		FogConnection struct {
		} `yaml:"fog_connection"`
	} `yaml:"buildpacks"`
	Packages struct {
		AppPackageDirectoryKey string `yaml:"app_package_directory_key"`
		FogAwsStorageOptions   struct {
		} `yaml:"fog_aws_storage_options"`
		FogConnection struct {
		} `yaml:"fog_connection"`
		MaxPackageSize         int `yaml:"max_package_size"`
		MaxValidPackagesStored int `yaml:"max_valid_packages_stored"`
	} `yaml:"packages"`
	Droplets struct {
		DropletDirectoryKey     string `yaml:"droplet_directory_key"`
		MaxStagedDropletsStored int    `yaml:"max_staged_droplets_stored"`
		FogAwsStorageOptions    struct {
		} `yaml:"fog_aws_storage_options"`
		FogConnection struct {
		} `yaml:"fog_connection"`
	} `yaml:"droplets"`
	RequestTimeoutInSeconds           int  `yaml:"request_timeout_in_seconds"`
	SkipCertVerify                    bool `yaml:"skip_cert_verify"`
	AppBitsUploadGracePeriodInSeconds int  `yaml:"app_bits_upload_grace_period_in_seconds"`
	SecurityEventLogging              struct {
		Enabled bool   `yaml:"enabled"`
		File    string `yaml:"file"`
	} `yaml:"security_event_logging"`
	BitsService struct {
		Enabled bool `yaml:"enabled"`
	} `yaml:"bits_service"`
	StatsdHost           string `yaml:"statsd_host"`
	StatsdPort           int    `yaml:"statsd_port"`
	CredentialReferences struct {
		InterpolateServiceBindings bool `yaml:"interpolate_service_bindings"`
	} `yaml:"credential_references"`
	DB struct {
		LogLevel                    string `yaml:"log_level"`
		MaxConnections              int    `yaml:"max_connections"`
		PoolTimeout                 int    `yaml:"pool_timeout"`
		LogDbQueries                bool   `yaml:"log_db_queries"`
		ConnectionValidationTimeout int    `yaml:"connection_validation_timeout"`
		Database                    string `yaml:"database"`
		DatabaseParts               struct {
			Adapter  string `yaml:"adapter"`
			Host     string `yaml:"host"`
			Port     int    `yaml:"port"`
			User     string `yaml:"user"`
			Password string `yaml:"password"`
			Database string `yaml:"database"`
		} `yaml:"database_parts"`
	} `yaml:"db"`
	Perm struct {
		Enabled               bool   `yaml:"enabled"`
		Hostname              string `yaml:"hostname"`
		Port                  int    `yaml:"port"`
		CACertPath            string `yaml:"ca_cert_path"`
		TimeoutInMilliseconds int    `yaml:"timeout_in_milliseconds"`
	} `yaml:"perm"`
}

func defaultConfig() *config {
	c := &config{}

	c.PidFilename = "/tmp/cloud_controller.pid"
	c.StacksFile = "config/stacks.yml"

	c.Logging.File = "/tmp/cloud_controller.log"
	c.Logging.Level = "debug"

	c.DB.MaxConnections = 42
	c.DB.DatabaseParts.Adapter = "postgres"
	c.DB.DatabaseParts.Host = "localhost"
	c.DB.DatabaseParts.Port = 5432
	c.DB.DatabaseParts.User = "postgres"
	c.DB.DatabaseParts.Database = "cc_test_integration_cc"

	c.Loggregator.Router = "127.0.0.1:3456"

	c.UAA.URL = "http://localhost:6789"
	c.UAA.InternalURL = "http://localhost:6789"
	c.UAA.ResourceID = "cloud_controller"
	c.UAA.SymmetricSecret = "tokensecret"
	c.UAA.CAFile = "spec/fixtures/certs/uaa_ca.crt"
	c.UAA.ClientTimeout = 60

	c.StatsdHost = "127.0.0.1"
	c.StatsdPort = 8125

	c.DefaultQuotaDefinition = "default"

	c.Renderer.MaxResultsPerPage = 100
	c.Renderer.DefaultResultsPerPage = 50
	c.Renderer.MaxInlineRelationsDepth = 3

	c.ExternalDomain = "capi.example.com"
	c.ExternalProtocol = "http"

	return c
}
