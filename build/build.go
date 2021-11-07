package build

import (
	"os"
	"strings"

	"github.com/jfrog/build-info-go/entities"
	"github.com/jfrog/build-info-go/utils"
)

type Build struct {
	buildName         string
	buildNumber       string
	projectKey        string
	tempDirPath       string
	logger            utils.Log
	agentName         string
	agentVersion      string
	buildAgentVersion string
	principal         string
	buildUrl          string
}

func NewBuild(buildName, buildNumber, projectKey, tempDirPath string, logger utils.Log) *Build {
	return &Build{
		buildName:   buildName,
		buildNumber: buildNumber,
		projectKey:  projectKey,
		tempDirPath: tempDirPath,
		logger:      logger,
	}
}

// This field is not saved in local cache. It is used only when creating a build-info using the ToBuildInfo() function.
func (b *Build) SetAgentName(agentName string) {
	b.agentName = agentName
}

// This field is not saved in local cache. It is used only when creating a build-info using the ToBuildInfo() function.
func (b *Build) SetAgentVersion(agentVersion string) {
	b.agentVersion = agentVersion
}

// This field is not saved in local cache. It is used only when creating a build-info using the ToBuildInfo() function.
func (b *Build) SetBuildAgentVersion(buildAgentVersion string) {
	b.buildAgentVersion = buildAgentVersion
}

// This field is not saved in local cache. It is used only when creating a build-info using the ToBuildInfo() function.
func (b *Build) SetPrincipal(principal string) {
	b.principal = principal
}

// This field is not saved in local cache. It is used only when creating a build-info using the ToBuildInfo() function.
func (b *Build) SetBuildUrl(buildUrl string) {
	b.buildUrl = buildUrl
}

// AddGoModule adds a Go module to this Build. Pass srcPath as an empty string if the root of the Go project is the working directory.
func (b *Build) AddGoModule(srcPath string) (*GoModule, error) {
	return newGoModule(srcPath, b)
}

// AddMavenModule adds a Maven module to this Build. Pass srcPath as an empty string if the root of the Maven project is the working directory.
func (b *Build) AddMavenModule(srcPath string) (*MavenModule, error) {
	return newMavenModule(b, srcPath)
}

// AddGradleModule adds a Gradle module to this Build. Pass srcPath as an empty string if the root of the Gradle project is the working directory.
func (b *Build) AddGradleModule(srcPath string) (*GradleModule, error) {
	return newGradleModule(b, srcPath)
}

func (b *Build) CollectEnv() error {
	envMap := make(map[string]string)
	for _, e := range os.Environ() {
		pair := strings.Split(e, "=")
		if len(pair[0]) != 0 {
			envMap["buildInfo.env."+pair[0]] = pair[1]
		}
	}
	partial := &entities.Partial{Env: envMap}
	return utils.SavePartialBuildInfo(b.buildName, b.buildNumber, b.projectKey, b.tempDirPath, partial, b.logger)
}

func (b *Build) ToBuildInfo() (*entities.BuildInfo, error) {
	buildInfo, err := utils.CreateBuildInfoFromPartials(b.buildName, b.buildNumber, b.projectKey, b.tempDirPath)
	if err != nil {
		return nil, err
	}
	buildInfo.SetAgentName(b.agentName)
	buildInfo.SetAgentVersion(b.agentVersion)
	buildInfo.SetBuildAgentVersion(b.buildAgentVersion)
	buildInfo.Principal = b.principal
	buildInfo.BuildUrl = b.buildUrl

	generatedBuildsInfo, err := utils.GetGeneratedBuildsInfo(b.buildName, b.buildNumber, b.projectKey, b.tempDirPath)
	if err != nil {
		return nil, err
	}
	for _, v := range generatedBuildsInfo {
		buildInfo.Append(v)
	}

	return buildInfo, nil
}

func (b *Build) Clean() error {
	return utils.RemoveBuildDir(b.buildName, b.buildNumber, b.projectKey, b.tempDirPath)
}

func generateEmptyBIFile(containingBuild *Build) (string, error) {
	buildDir, err := utils.CreateTempBuildFile(containingBuild.buildName, containingBuild.buildNumber, containingBuild.projectKey, containingBuild.tempDirPath)
	if err != nil {
		return "", err
	}
	if err := buildDir.Close(); err != nil {
		return "", err
	}
	// If this is a Windows machine there is a need to modify the path for the build info file to match Java syntax with double \\
	return utils.DoubleWinPathSeparator(buildDir.Name()), nil
}
