package smokescreen

import (
	"errors"
	"io/ioutil"
	"os"
	"strconv"
	"time"

	"gopkg.in/yaml.v2"
)

type yamlConfigTls struct {
	CertFile      string   `yaml:"cert_file"`
	KeyFile       string   `yaml:"key_file"`
	ClientCAFiles []string `yaml:"client_ca_files"`
	CRLFiles      []string `yaml:"crl_files"`
}

// Port and ExitTimeout use a pointer so we can distinguish unset vs explicit
// zero, to avoid overriding a non-zero default when the value is not set.
type yamlConfig struct {
	Ip                   string
	Port                 *uint16
	DenyRanges           []string       `yaml:"deny_ranges"`
	AllowRanges          []string       `yaml:"allow_ranges"`
	Resolvers            []string       `yaml:"resolver_addresses"`
	ConnectTimeout       time.Duration  `yaml:"connect_timeout"`
	ExitTimeout          *time.Duration `yaml:"exit_timeout"`
	StatsdAddress        string         `yaml:"statsd_address"`
	EgressAclFile        string         `yaml:"acl_file"`
	SupportProxyProtocol bool           `yaml:"support_proxy_protocol"`
	DenyMessageExtra     string         `yaml:"deny_message_extra"`
	AllowMissingRole     bool           `yaml:"allow_missing_role"`

	StatsSocketDir      string `yaml:"stats_socket_dir"`
	StatsSocketFileMode string `yaml:"stats_socket_file_mode"`

	Tls *yamlConfigTls

	// Currently not configurable via YAML: RoleFromRequest, Log, DisabledAclPolicyActions
}

func (c *Config) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var yc yamlConfig
	*c = *NewConfig()

	err := unmarshal(&yc)
	if err != nil {
		return err
	}

	c.Ip = yc.Ip

	if yc.Port != nil {
		c.Port = *yc.Port
	}

	err = c.SetDenyRanges(yc.DenyRanges)
	if err != nil {
		return err
	}

	err = c.SetAllowRanges(yc.AllowRanges)
	if err != nil {
		return err
	}

	err = c.SetResolverAddresses(yc.Resolvers)
	if err != nil {
		return err
	}

	c.ConnectTimeout = yc.ConnectTimeout
	if yc.ExitTimeout != nil {
		c.ExitTimeout = *yc.ExitTimeout
	}

	err = c.SetupStatsd(yc.StatsdAddress)
	if err != nil {
		return err
	}

	if yc.EgressAclFile != "" {
		err = c.SetupEgressAcl(yc.EgressAclFile)
		if err != nil {
			return err
		}
	}

	c.SupportProxyProtocol = yc.SupportProxyProtocol

	if yc.StatsSocketDir != "" {
		c.StatsSocketDir = yc.StatsSocketDir
	}

	if yc.StatsSocketFileMode != "" {
		filemode, err := strconv.ParseInt(yc.StatsSocketFileMode, 8, 9)

		if err != nil {
			c.Log.Fatal(err)
		}

		c.StatsSocketFileMode = os.FileMode(filemode)
	}

	if yc.Tls != nil {
		if yc.Tls.CertFile == "" {
			return errors.New("'tls' section requires 'cert_file'")
		}

		key_file := yc.Tls.KeyFile
		if key_file == "" {
			// Assume CertFile is a cert+key bundle
			key_file = yc.Tls.CertFile
		}

		err = c.SetupTls(yc.Tls.CertFile, key_file, yc.Tls.ClientCAFiles)
		if err != nil {
			return err
		}

		c.SetupCrls(yc.Tls.CRLFiles)
	}

	c.AllowMissingRole = yc.AllowMissingRole
	c.AdditionalErrorMessageOnDeny = yc.DenyMessageExtra

	return nil
}

func LoadConfig(filePath string) (*Config, error) {
	bytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	config := &Config{}
	if err := yaml.UnmarshalStrict(bytes, config); err != nil {
		return nil, err
	}

	return config, nil
}
