package bosh

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"

	"github.com/EngineerBetter/control-tower/bosh/internal/boshcli"
	"github.com/EngineerBetter/control-tower/bosh/internal/workingdir"
	"github.com/EngineerBetter/control-tower/config"
	"github.com/EngineerBetter/control-tower/db"
	"github.com/EngineerBetter/control-tower/iaas"
	"github.com/EngineerBetter/control-tower/terraform"
	"github.com/EngineerBetter/control-tower/util"
	"gopkg.in/yaml.v2"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

// StateFilename is default name for bosh-init state file
const StateFilename = "director-state.json"

// CredsFilename is default name for bosh-init creds file
const CredsFilename = "director-creds.yml"

// IClient is a client for performing bosh-init commands
//
//counterfeiter:generate . IClient
type IClient interface {
	Deploy([]byte, []byte, bool) ([]byte, []byte, error)
	Cleanup() error
	Instances() ([]Instance, error)
	CreateEnv([]byte, []byte, string) ([]byte, []byte, error)
	Recreate() error
	Locks() ([]byte, error)
}

// Instance represents a vm deployed by BOSH
type Instance struct {
	Name  string
	IP    string
	State string
}

// ClientFactory creates a new IClient
type ClientFactory func(config config.ConfigView, outputs terraform.Outputs, stdout, stderr io.Writer, provider iaas.Provider, versionFile []byte) (IClient, error)

// New returns an IAAS specific implementation of BOSH client
func New(config config.ConfigView, outputs terraform.Outputs, stdout, stderr io.Writer, provider iaas.Provider, versionFile []byte) (IClient, error) {
	workingdir, err := workingdir.New()
	if err != nil {
		return nil, err
	}

	var binaries map[string]util.BinaryPaths

	err = json.Unmarshal(versionFile, &binaries)
	if err != nil {
		return nil, err
	}

	boshCLIPath, err := util.DownloadBOSHCLI(binaries)
	if err != nil {
		return nil, fmt.Errorf("failed to determine BOSH CLI path: [%v]", err)
	}

	boshCLI := boshcli.New(boshCLIPath, exec.Command)

	switch provider.IAAS() {
	case iaas.AWS:
		return NewAWSClient(config, outputs, workingdir, stdout, stderr, provider, boshCLI, versionFile)
	case iaas.GCP:
		return NewGCPClient(config, outputs, workingdir, stdout, stderr, provider, boshCLI, versionFile)
	}
	return nil, fmt.Errorf("IAAS not supported: %s", provider.IAAS())
}

func instances(boshCLI boshcli.ICLI, ip, password, ca string) ([]Instance, error) {
	output := new(bytes.Buffer)

	if err := boshCLI.RunAuthenticatedCommand(
		"instances",
		ip,
		password,
		ca,
		false,
		output,
		"--json",
	); err != nil {
		return nil, fmt.Errorf("Error [%s] running `bosh instances`. stdout: [%s]", err, output.String())
	}

	jsonOutput := struct {
		Tables []struct {
			Rows []struct {
				Instance     string `json:"instance"`
				IPs          string `json:"ips"`
				ProcessState string `json:"process_state"`
			} `json:"Rows"`
		} `json:"Tables"`
	}{}

	if err := json.NewDecoder(output).Decode(&jsonOutput); err != nil {
		return nil, err
	}

	instances := []Instance{}

	for _, table := range jsonOutput.Tables {
		for _, row := range table.Rows {
			instances = append(instances, Instance{
				Name:  row.Instance,
				IP:    row.IPs,
				State: row.ProcessState,
			})
		}
	}

	return instances, nil
}

type external_tls_config struct {
	External_tls external_tls `yaml:"external_tls"`
}

type external_tls struct {
	Certificate string `yaml:"certificate"`
	Private_key string `yaml:"private_key"`
}

func saveFilesToWorkingDir(workingdir workingdir.IClient, provider iaas.Provider, creds []byte, external_tls_certificate string, external_tls_private_key string) error {
	concourseVersionsContents, _ := provider.Choose(iaas.Choice{
		AWS: awsConcourseVersions,
		GCP: gcpConcourseVersions,
	}).([]byte)
	concourseSHAsContents, _ := provider.Choose(iaas.Choice{
		AWS: awsConcourseSHAs,
		GCP: gcpConcourseSHAs,
	}).([]byte)

	// write to yaml file external_tls.certificate, keys are external_tls.certificate external_tls.private_key
	external_tls_config := external_tls_config{
		External_tls: external_tls{
			Certificate: external_tls_certificate,
			Private_key: external_tls_private_key,
		},
	}
	external_tls_config_yaml, err := yaml.Marshal(external_tls_config)
	if err != nil {
		return err
	}

	filesToSave := map[string][]byte{
		concourseVersionsFilename:             concourseVersionsContents,
		concourseSHAsFilename:                 concourseSHAsContents,
		concourseManifestFilename:             concourseManifestContents,
		concourseGrafanaFilename:              concourseGrafana,
		concourseBitBucketAuthFilename:        concourseBitBucketAuth,
		concourseGitHubAuthFilename:           concourseGitHubAuth,
		concourseGitHubEnterpriseAuthFilename: concourseGithubEnterpriseAuth,
		concourseMainGitHubAuthFilename:       concourseMainGitHubAuth,
		concourseMicrosoftAuthFilename:        concourseMicrosoftAuth,
		concourseEphemeralWorkersFilename:     concourseEphemeralWorkers,
		concourseNoMetricsFilename:            concourseNoMetrics,
		credsFilename:                         creds,
		extraTagsFilename:                     extraTags,
		psqlCAFilename:                        []byte(db.RDSRootCert),
		concourseCertFilename:                 external_tls_config_yaml,
	}

	for filename, contents := range filesToSave {
		_, err := workingdir.SaveFileToWorkingDir(filename, contents)
		if err != nil {
			return fmt.Errorf("failed to save %s to working directory: [%v]", filename, err)
		}
	}
	return nil
}
