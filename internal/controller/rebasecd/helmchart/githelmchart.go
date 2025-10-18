package helmchart

import (
	"path"
	"regexp"
	"strings"

	rebasecd "github.com/bensoer/rebasecd/internal/controller/rebasecd"
	"github.com/bensoer/rebasecd/internal/controller/utils"
	"helm.sh/helm/v3/pkg/action"

	"fmt"
	"os"
	"path/filepath"

	git "github.com/go-git/go-git/v5"
)

type GitHelmChart struct {
	// HTTPs or SSH URL of the git repository
	gitRepositoryUrl string
	// Path within the git repository where the Helm chart is located
	chartRootPath string

	repoTempDir string
	*HelmChart
}

func NewGitHelmChart(gitRepositoryUrl, chartRootPath, releaseName, namespace string, valuesFiles []string) rebasecd.ChartHandler {
	return &GitHelmChart{
		chartRootPath:    chartRootPath,
		gitRepositoryUrl: gitRepositoryUrl,
		HelmChart: &HelmChart{
			ReleaseName: releaseName,
			ValuesFiles: valuesFiles,
			Namespace:   namespace,
		},
	}
}

/**
 * GitHelmchart.GetChart retrieves the Helm chart from a Git repository
 */
func (ghc *GitHelmChart) GetChart() error {

	// We are trying to resolve getting a Helm chart located in a git repository

	// Git clone the repository into a temp directory
	tempDir, err := ghc.cloneRepoToTemp(ghc.gitRepositoryUrl)
	if err != nil {
		return fmt.Errorf("failed to clone git repo: %w", err)
	}
	ghc.repoTempDir = tempDir

	actionConfig, settings, err := utils.NewActionConfigAndSettings(ghc.Namespace)
	if err != nil {
		return fmt.Errorf("failed to create action config: %w", err)
	}

	chartPathOpts := action.ChartPathOptions{}
	chartPath, err := chartPathOpts.LocateChart(filepath.Join(tempDir, ghc.chartRootPath), settings)
	if err != nil {
		return fmt.Errorf("failed to locate chart: %w", err)
	}

	// Configure HelmChart with this information
	ghc.HelmChart.ActionConfig = actionConfig
	ghc.HelmChart.Settings = settings
	ghc.HelmChart.ChartPath = chartPath

	return nil

}

func (ghc *GitHelmChart) Cleanup() {
	if ghc.repoTempDir != "" {
		os.RemoveAll(ghc.repoTempDir)
	}
	ghc.HelmChart.Cleanup()
}

func (ghc *GitHelmChart) cloneRepoToTemp(repoURL string) (string, error) {
	repoName := ghc.repoNameFromURL(repoURL)

	// Create a temporary directory for the clone
	tempDir, err := os.MkdirTemp("", fmt.Sprintf("%s-*", repoName))
	if err != nil {
		return "", fmt.Errorf("failed to create temp dir: %w", err)
	}

	// Clone the repository into that directory
	fmt.Printf("Cloning %s into %s...\n", repoURL, tempDir)
	_, err = git.PlainClone(tempDir, false, &git.CloneOptions{
		URL:      repoURL,
		Progress: os.Stdout,
	})
	if err != nil {
		os.RemoveAll(tempDir) // cleanup on failure
		return "", fmt.Errorf("failed to clone repo: %w", err)
	}

	return filepath.Clean(tempDir), nil
}

func (ghc *GitHelmChart) repoNameFromURL(repoURL string) string {
	// Remove trailing slashes or .git suffix
	repoURL = strings.TrimSuffix(repoURL, "/")
	repoURL = strings.TrimSuffix(repoURL, ".git")

	// Handle SSH URLs like: git@github.com:user/repo
	if strings.Contains(repoURL, ":") && strings.Contains(repoURL, "@") {
		// Replace first ":" after the host with "/" to make it look like a normal path
		re := regexp.MustCompile(`^[^:]+:[^/].*$`)
		if re.MatchString(repoURL) {
			parts := strings.SplitN(repoURL, ":", 2)
			repoURL = parts[0] + "/" + parts[1]
		}
	}

	// Extract the last element of the path
	repoName := path.Base(repoURL)
	return repoName
}
