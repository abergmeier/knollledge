package config

import "github.com/abergmeier/knollledge/internal/job"

var (
	predefinedCodeSearches = map[string]job.MakeCodeSearchFunc{
		"bazel-package":           job.MakeBazelPackageCodeSearch,
		"buf-configuration":       job.MakeBufConfigurationCodeSearch,
		"cargo-configuration":     job.MakeCargoConfigurationCodeSearch,
		"container-configuration": job.MakeContainerConfigurationCodeSearch,
		"go-modules":              job.MakeGoModuleCodeSearch,
		"poetry-configuration":    job.MakePoetryConfigurationCodeSearch,
		"protobuf-definition":     job.MakeProtobufDefinitionCodeSearch,
		"skaffold-configuration":  job.MakeSkaffoldConfigurationCodeSearch,
		"terraform-backend":       job.MakeTerraformBackendCodeSearch,
		"terraform-gke-cluster":   job.MakeTerraformGKECluster,
	}
)
