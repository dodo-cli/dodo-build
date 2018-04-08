package image

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/builder/dockerignore"
	"github.com/docker/docker/pkg/archive"
	"github.com/docker/docker/pkg/fileutils"
	"github.com/docker/docker/pkg/idtools"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/docker/docker/pkg/term"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
)

func build(ctx context.Context, options Options) (string, error) {
	args := map[string]*string{}
	for _, arg := range options.Build.Args {
		switch values := strings.SplitN(arg, "=", 2); len(values) {
		case 1:
			args[values[0]] = nil
		case 2:
			args[values[0]] = &values[1]
		}
	}

	contextDir, err := getContextDir(options.Build.Context)
	if err != nil {
		return "", err
	}
	dockerfile, err := getDockerfile(options.Build.Dockerfile, contextDir)
	if err != nil {
		return "", err
	}
	excludes, err := getDockerignore(contextDir, dockerfile)
	if err != nil {
		return "", err
	}

	// TODO: validate that all files in the context are ok
	tarStream, err := archive.TarWithOptions(contextDir, &archive.TarOptions{
		ExcludePatterns: excludes,
		ChownOpts:       &idtools.IDPair{UID: 0, GID: 0},
	})
	if err != nil {
		return "", err
	}

	response, err := options.Client.ImageBuild(
		ctx,
		tarStream,
		types.ImageBuildOptions{
			SuppressOutput: false, // TODO: quiet mode
			NoCache:        options.Build.NoCache,
			Remove:         true,
			ForceRemove:    true,
			PullParent:     options.ForcePull,
			Dockerfile:     options.Build.Dockerfile,
			BuildArgs:      args,
		},
	)
	if err != nil {
		return "", err
	}
	defer func() {
		if err := response.Body.Close(); err != nil {
			log.Error(err)
		}
	}()

	name := ""
	aux := func(auxJSON *json.RawMessage) {
		var result types.BuildResult
		if err := json.Unmarshal(*auxJSON, &result); err == nil {
			name = result.ID
		} else {
			log.Error(err)
		}
	}

	outFd, isTerminal := term.GetFdInfo(os.Stdout)
	err = jsonmessage.DisplayJSONMessagesStream(response.Body, os.Stdout, outFd, isTerminal, aux)
	if err != nil {
		return "", err
	}
	if name == "" {
		return "", errors.New("build complete, but the server did not send an image id")
	}
	return name, nil
}

func getContextDir(givenContext string) (string, error) {
	contextDir := givenContext
	if contextDir == "" {
		contextDir = "."
	}
	contextDir, err := filepath.Abs(contextDir)
	if err != nil {
		return "", err
	}
	contextDir, err = filepath.EvalSymlinks(contextDir)
	if err != nil {
		return "", err
	}
	stat, err := os.Lstat(contextDir)
	if err != nil {
		return "", err
	}
	if !stat.IsDir() {
		return "", fmt.Errorf("context must be a directory: %s", contextDir)
	}
	return contextDir, nil
}

func getDockerfile(givenDockerfile string, contextDir string) (string, error) {
	dockerfile := givenDockerfile
	if dockerfile == "" {
		dockerfile = filepath.Join(contextDir, "Dockerfile")
	}
	if !filepath.IsAbs(dockerfile) {
		dockerfile = filepath.Join(contextDir, dockerfile)
	}
	dockerfile, err := filepath.EvalSymlinks(dockerfile)
	if err != nil {
		return "", err
	}
	_, err = os.Lstat(dockerfile)
	if err != nil {
		return "", err
	}
	dockerfile, err = filepath.Rel(contextDir, dockerfile)
	if err != nil {
		return "", err
	}
	dockerfile, err = archive.CanonicalTarNameForPath(dockerfile)
	if err != nil {
		return "", err
	}
	return dockerfile, nil
}

func getDockerignore(contextDir string, dockerfile string) ([]string, error) {
	file, err := os.Open(filepath.Join(contextDir, ".dockerignore"))
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Error(err)
		}
	}()

	excludes, err := dockerignore.ReadAll(file)
	if err != nil {
		return nil, err
	}

	if keep, _ := fileutils.Matches(".dockerignore", excludes); keep {
		excludes = append(excludes, "!.dockerignore")
	}
	if keep, _ := fileutils.Matches(dockerfile, excludes); keep {
		excludes = append(excludes, "!"+dockerfile)
	}

	return excludes, nil
}