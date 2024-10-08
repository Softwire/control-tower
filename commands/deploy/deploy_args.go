package deploy

import (
	"encoding/pem"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/asaskevich/govalidator"
	"gopkg.in/urfave/cli.v1"
)

// Args are arguments passed to the deploy command
type Args struct {
	IAAS                string
	IAASIsSet           bool
	Region              string
	RegionIsSet         bool
	Domain              string
	DomainIsSet         bool
	TLSCert             string
	TLSCertIsSet        bool
	TLSKey              string
	TLSKeyIsSet         bool
	WorkerCount         int
	WorkerCountIsSet    bool
	WorkerSize          string
	WorkerSizeIsSet     bool
	WebSize             string
	WebSizeIsSet        bool
	PersistentDiskSize  string
	PersistentDiskIsSet bool
	SelfUpdate          bool
	SelfUpdateIsSet     bool
	DBSize              string
	// DBSizeIsSet is true if the user has manually specified the db-size (ie, it's not the default)
	DBSizeIsSet                    bool
	RDSDiskEncryption              bool
	RDSDiskEncryptionIsSet         bool
	EnableGlobalResources          bool
	EnableGlobalResourcesIsSet     bool
	EnablePipelineInstances        bool
	EnablePipelineInstancesIsSet   bool
	InfluxDbRetention              string
	InfluxDbRetentionIsSet         bool
	Namespace                      string
	NamespaceIsSet                 bool
	AllowIPs                       string
	AllowIPsIsSet                  bool
	BitbucketAuthClientID          string
	BitbucketAuthClientIDIsSet     bool
	BitbucketAuthClientSecret      string
	BitbucketAuthClientSecretIsSet bool
	// BitbucketAuthIsSet is true if the user has specified both the --bitbucket-auth-client-secret and --bitbucket-auth-client-id flags
	BitbucketAuthIsSet          bool
	GithubAuthClientID          string
	GithubAuthClientIDIsSet     bool
	GithubAuthClientSecret      string
	GithubAuthClientSecretIsSet bool
	// GithubAuthIsSet is true if the user has specified both the --github-auth-client-secret and --github-auth-client-id flags
	GithubAuthIsSet           bool
	GithubAuthHost            string
	GithubAuthHostIsSet       bool
	GithubAuthCaCert          string
	GithubAuthCaCertIsSet     bool
	GithubEnterpriseAuthIsSet bool
	// GithubEnterpriseAuthIsSet is true if the user has specified both the --github-auth-host and --github-auth-ca-cert flags
	MainGithubUsers      string
	MainGithubUsersIsSet bool
	MainGithubTeams      string
	MainGithubTeamsIsSet bool
	MainGithubOrgs       string
	MainGithubOrgsIsSet  bool
	// MainGithubAuthIsSet is true if any main team github auth flags have been used
	MainGithubAuthIsSet            bool
	MicrosoftAuthClientID          string
	MicrosoftAuthClientIDIsSet     bool
	MicrosoftAuthClientSecret      string
	MicrosoftAuthClientSecretIsSet bool
	MicrosoftAuthTenant            string
	MicrosoftAuthTenantIsSet       bool
	// MicrosoftAuthIsSet is true if the user has specified both the --microsoft-auth-client-secret and --microsoft-auth-client-id flags
	MicrosoftAuthIsSet bool
	NoMetrics          bool
	NoMetricsIsSet     bool
	Tags               cli.StringSlice
	// TagsIsSet is true if the user has specified tags using --add-tag
	TagsIsSet        bool
	Spot             bool
	SpotIsSet        bool
	Zone             string
	ZoneIsSet        bool
	WorkerType       string
	WorkerTypeIsSet  bool
	NetworkCIDR      string
	NetworkCIDRIsSet bool
	PublicCIDR       string
	PublicCIDRIsSet  bool
	PrivateCIDR      string
	PrivateCIDRIsSet bool
	RDS1CIDR         string
	RDS1CIDRIsSet    bool
	RDS2CIDR         string
	RDS2CIDRIsSet    bool
}

// MarkSetFlags is marking the IsSet DeployArgs
func (a *Args) MarkSetFlags(c FlagSetChecker) error {
	for _, f := range c.FlagNames() {
		if c.IsSet(f) {
			switch f {
			case "region":
				a.RegionIsSet = true
			case "enable-global-resources":
				a.EnableGlobalResourcesIsSet = true
			case "enable-pipeline-instances":
				a.EnablePipelineInstancesIsSet = true
			case "influxdb-retention-period":
				a.InfluxDbRetentionIsSet = true
			case "domain":
				a.DomainIsSet = true
			case "tls-cert":
				a.TLSCertIsSet = true
			case "tls-key":
				a.TLSKeyIsSet = true
			case "workers":
				a.WorkerCountIsSet = true
			case "worker-size":
				a.WorkerSizeIsSet = true
			case "web-size":
				a.WebSizeIsSet = true
			case "persistent-disk":
				a.PersistentDiskIsSet = true
			case "iaas":
				a.IAASIsSet = true
			case "self-update":
				a.SelfUpdateIsSet = true
			case "db-size":
				a.DBSizeIsSet = true
			case "rds-disk-encryption":
				a.RDSDiskEncryptionIsSet = true
			case "spot", "preemptible":
				a.SpotIsSet = true
			case "allow-ips":
				a.AllowIPsIsSet = true
			case "bitbucket-auth-client-id":
				a.BitbucketAuthClientIDIsSet = true
			case "bitbucket-auth-client-secret":
				a.BitbucketAuthClientSecretIsSet = true
			case "github-auth-client-id":
				a.GithubAuthClientIDIsSet = true
			case "github-auth-client-secret":
				a.GithubAuthClientSecretIsSet = true
			case "github-auth-host":
				a.GithubAuthHostIsSet = true
			case "github-auth-ca-cert":
				a.GithubAuthCaCertIsSet = true
			case "main-team-github-users":
				a.MainGithubUsersIsSet = true
			case "main-team-github-teams":
				a.MainGithubTeamsIsSet = true
			case "main-team-github-orgs":
				a.MainGithubOrgsIsSet = true
			case "microsoft-auth-client-id":
				a.MicrosoftAuthClientIDIsSet = true
			case "microsoft-auth-client-secret":
				a.MicrosoftAuthClientSecretIsSet = true
			case "microsoft-auth-tenant":
				a.MicrosoftAuthTenantIsSet = true
			case "add-tag":
				a.TagsIsSet = true
			case "namespace":
				a.NamespaceIsSet = true
			case "zone":
				a.ZoneIsSet = true
			case "worker-type":
				a.WorkerTypeIsSet = true
			case "vpc-network-range":
				a.NetworkCIDRIsSet = true
			case "public-subnet-range":
				a.PublicCIDRIsSet = true
			case "private-subnet-range":
				a.PrivateCIDRIsSet = true
			case "rds-subnet-range1":
				a.RDS1CIDRIsSet = true
			case "rds-subnet-range2":
				a.RDS2CIDRIsSet = true
			case "no-metrics":
				a.NoMetricsIsSet = true
			default:
				return fmt.Errorf("flag %q is not supported by deployment flags", f)
			}
		}
	}
	a.BitbucketAuthIsSet = c.IsSet("bitbucket-auth-client-id") && c.IsSet("bitbucket-auth-client-secret")
	a.GithubAuthIsSet = c.IsSet("github-auth-client-id") && c.IsSet("github-auth-client-secret")
	a.GithubEnterpriseAuthIsSet = c.IsSet("github-auth-host") && c.IsSet("github-auth-ca-cert")
	a.MicrosoftAuthIsSet = c.IsSet("microsoft-auth-client-id") && c.IsSet("microsoft-auth-client-secret")
	a.MainGithubAuthIsSet = c.IsSet("main-team-github-users") || c.IsSet("main-team-github-teams") || c.IsSet("main-team-github-orgs")

	return nil
}

// WorkerSizes are the permitted concourse worker sizes
var WorkerSizes = []string{"medium", "large", "xlarge", "2xlarge", "4xlarge", "12xlarge", "24xlarge"}

// WebSizes are the permitted concourse web sizes
var WebSizes = []string{"small", "medium", "large", "xlarge", "2xlarge"}

// PersistentDiskSizes are the permitted concourse persistent disk sizes
var PersistentDiskSizes = []string{"small", "default", "medium", "large"}

// AllowedDBSizes contains the valid values for --db-size flag
var AllowedDBSizes = []string{"small", "medium", "large", "xlarge", "2xlarge", "4xlarge"}

// Validate validates that flag interdependencies
func (a Args) Validate() error {
	if !a.IAASIsSet {
		return fmt.Errorf("--iaas flag not set")
	}

	if err := a.validateCertFields(); err != nil {
		return err
	}

	if err := a.validateWorkerFields(); err != nil {
		return err
	}

	if err := a.validateWebFields(); err != nil {
		return err
	}

	if err := a.validatePersistentDiskFields(); err != nil {
		return err
	}

	if err := a.validateDBFields(); err != nil {
		return err
	}

	if err := a.validateGithubFields(); err != nil {
		return err
	}

	if err := a.validateNetworkRanges(); err != nil {
		return err
	}

	if err := a.validateTags(); err != nil {
		return err
	}

	if a.MainGithubAuthIsSet {
		if err := a.validateMainAuth(); err != nil {
			return err
		}
	}

	return nil
}

func (a Args) validateCertFields() error {
	if a.TLSKey != "" && a.TLSCert == "" {
		return errors.New("--tls-key requires --tls-cert to also be provided")
	}
	if a.TLSCert != "" && a.TLSKey == "" {
		return errors.New("--tls-cert requires --tls-key to also be provided")
	}
	if (a.TLSKey != "" || a.TLSCert != "") && a.Domain == "" {
		return errors.New("custom certificates require --domain to be provided")
	}

	return nil
}

func (a Args) validateWorkerFields() error {

	if a.WorkerCount < 1 {
		return errors.New("minimum number of workers is 1")
	}

	if a.WorkerTypeIsSet && strings.ToLower(a.IAAS) != "aws" {
		return errors.New("worker-type is only defined on AWS")
	}

	re := regexp.MustCompile("^m5$|^m5a$|^m4$")
	if a.WorkerTypeIsSet && !re.MatchString(a.WorkerType) {
		return fmt.Errorf("worker-type %s is invalid: must be one of m4, m5, or m5a", a.WorkerType)
	}

	for _, size := range WorkerSizes {
		if size == a.WorkerSize {
			return nil
		}
	}
	return fmt.Errorf("unknown worker size: `%s`. Valid sizes are: %v", a.WorkerSize, WorkerSizes)
}

func (a Args) validateWebFields() error {
	if a.NoMetricsIsSet && a.InfluxDbRetentionIsSet {
		return fmt.Errorf("no-metrics is invalid when used with influxdb-retention-period")
	}

	for _, size := range WebSizes {
		if size == a.WebSize {
			return nil
		}
	}
	return fmt.Errorf("unknown web node size: `%s`. Valid sizes are: %v", a.WebSize, WebSizes)
}

func (a Args) validatePersistentDiskFields() error {
	for _, size := range PersistentDiskSizes {
		if size == a.PersistentDiskSize {
			return nil
		}
	}
	return fmt.Errorf("unknown persistent disk size: `%s`. Valid sizes are: %v", a.PersistentDiskSize, PersistentDiskSizes)
}

func (a Args) validateDBFields() error {
	for _, size := range AllowedDBSizes {
		if size == a.DBSize {
			return nil
		}
	}
	return fmt.Errorf("unknown DB size: `%s`. Valid sizes are: %v", a.DBSize, AllowedDBSizes)
}

func (a Args) validateGithubFields() error {
	if a.GithubAuthClientID != "" && a.GithubAuthClientSecret == "" {
		return errors.New("--github-auth-client-id requires --github-auth-client-secret to also be provided")
	}
	if a.GithubAuthClientID == "" && a.GithubAuthClientSecret != "" {
		return errors.New("--github-auth-client-secret requires --github-auth-client-id to also be provided")
	}
	if a.GithubAuthHost != "" && (!a.GithubAuthIsSet || a.GithubAuthCaCert == "") {
		return errors.New("--github-auth-host requires --github-auth-ca-cert, --github-auth-client-id, --github-auth-client-secret to also be provided")
	}
	if a.GithubAuthCaCert != "" && (!a.GithubAuthIsSet || a.GithubAuthHost == "") {
		return errors.New("--github-auth-ca-cert requires --github-auth-host, --github-auth-client-id, --github-auth-client-secret to also be provided")
	}
	if a.GithubEnterpriseAuthIsSet && !a.certParseable() {
		return errors.New("unable to decode value passed to --github-auth-ca-cert. Provide a CA certificate in PEM format")
	}
	if a.GithubEnterpriseAuthIsSet && !govalidator.IsDNSName(a.GithubAuthHost) {
		return errors.New("--github-auth-host must be a valid DNS address (omitting protocol)")
	}
	return nil
}

func (a Args) certParseable() bool {
	decodedCert, _ := pem.Decode([]byte(a.GithubAuthCaCert))
	return decodedCert != nil
}

func (a Args) validateNetworkRanges() error {
	if a.PublicCIDR != "" || a.PrivateCIDR != "" {
		if a.PublicCIDR == "" || a.PrivateCIDR == "" {
			return errors.New("both --public-subnet-range and --private-subnet-range are required when either is provided")
		}
	}

	return nil
}

func (a Args) validateTags() error {
	pattern := regexp.MustCompile(`\w+=\w+`)
	for _, tag := range a.Tags {
		if !pattern.MatchString(tag) {
			return fmt.Errorf("`%v` is not in the format `key=value`", tag)
		}
	}
	return nil
}

func (a Args) validateMainAuth() error {
	if err := a.validateMainAuthFlags(); err != nil {
		return err
	}
	return nil
}

// This is a partial validation check based on the format expected by Concourse:
// https://concourse-ci.org/github-auth.html#configuring-main-team-authorization
// as well as some of the validation rules imposed by GitHub:
// https://github.com/dead-claudia/github-limits
// We don't check all rules (such as reserved names) as we're just trying to save
// users from typos rather than from creating invalid entities in GitHub
func (a Args) validateMainAuthFlags() error {
	userAndOrgRe := regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9-]*$`)
	teamRe := regexp.MustCompile(`^[a-zA-Z0-9-]+$`)

	if a.MainGithubUsersIsSet {
		for _, user := range strings.Split(a.MainGithubUsers, ",") {
			if userAndOrgRe.FindString(strings.TrimSpace(user)) == "" {
				return fmt.Errorf("Invalid user %q provided to --main-team-github-users", user)
			}
		}
	}

	if a.MainGithubOrgsIsSet {
		for _, org := range strings.Split(a.MainGithubOrgs, ",") {
			if userAndOrgRe.FindString(strings.TrimSpace(org)) == "" {
				return fmt.Errorf("Invalid org %q provided to --main-team-github-orgs", org)
			}
		}
	}

	if a.MainGithubTeamsIsSet {
		for _, orgTeam := range strings.Split(a.MainGithubTeams, ",") {
			parts := strings.Split(orgTeam, ":")
			org := strings.TrimSpace(parts[0])
			if len(parts) != 2 {
				return fmt.Errorf("Invalid team %q does not contain org", org)
			}
			team := strings.TrimSpace(parts[1])
			if userAndOrgRe.FindString(org) == "" {
				return fmt.Errorf("Invalid org %q provided for team %q in --main-team-github-teams", org, team)
			}
			if teamRe.FindString(team) == "" {
				return fmt.Errorf("Invalid team %q provided to --main-team-github-teams", orgTeam)
			}
		}
	}
	return nil
}

// FlagSetChecker allows us to find out if flags were set, and what the names of all flags are
type FlagSetChecker interface {
	IsSet(name string) bool
	FlagNames() (names []string)
}

// ContextWrapper wraps a CLI context for testing
type ContextWrapper struct {
	c *cli.Context
}

// IsSet tells you if a user provided a flag
func (t *ContextWrapper) IsSet(name string) bool {
	return t.c.IsSet(name)
}

// FlagNames lists all flags it's possible for a user to provide
func (t *ContextWrapper) FlagNames() (names []string) {
	return t.c.FlagNames()
}
