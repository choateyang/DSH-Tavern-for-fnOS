package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const (
	nodeBinDir        = "/var/apps/nodejs_v24/target/bin"
	tavernRepo        = "flizzywine/dsh-tavern"
	tavernCommitURL   = "https://api.github.com/repos/" + tavernRepo + "/commits/main"
	tavernPnpmVersion = "11.25.0"
)

func findBin(preferred, fallback string) string {
	if _, err := os.Stat(preferred); err == nil {
		return preferred
	}
	if p, err := exec.LookPath(fallback); err == nil {
		return p
	}
	return preferred
}

func nodeBin() string { return findBin(filepath.Join(nodeBinDir, "node"), "node") }
func npmBin() string  { return findBin(filepath.Join(nodeBinDir, "npm"), "npm") }
func pnpmBin() string { return filepath.Join(globalPnpmDir, "node_modules", ".bin", "pnpm") }

// withEnv replaces keys instead of appending duplicate environment entries.
func withEnv(base []string, updates map[string]string) []string {
	result := make([]string, 0, len(base)+len(updates))
	for _, entry := range base {
		idx := strings.IndexByte(entry, '=')
		if idx > 0 {
			if _, exists := updates[entry[:idx]]; exists {
				continue
			}
		}
		result = append(result, entry)
	}
	for key, value := range updates {
		result = append(result, key+"="+value)
	}
	return result
}

func tavernEnv(extra map[string]string) []string {
	cfg := GetConfig()
	pathValue := os.Getenv("PATH")
	privateBin := filepath.Dir(pnpmBin())
	if pathValue == "" {
		pathValue = privateBin
	} else if !strings.Contains(string(os.PathListSeparator)+pathValue+string(os.PathListSeparator), string(os.PathListSeparator)+privateBin+string(os.PathListSeparator)) {
		pathValue = privateBin + string(os.PathListSeparator) + pathValue
	}
	updates := map[string]string{
		"DSH_HOME":                globalDshHome,
		"DSH_TAVERN_CLI_HOME":     globalDshHome,
		"DSH_TAVERN_RUNTIME_HOST": "cli",
		"DSH_TAVERN_HOST":         "cli",
		"DSH_TAVERN_PORT":         fmt.Sprintf("%d", cfg.GetServerPort()),
		"DSH_TAVERN_NPM_REGISTRY": cfg.GetNpmRegistry(),
		"DSH_TAVERN_NO_OPEN":      "1",
		"DSH_TAVERN_NO_START":     "1",
		"PATH":                    pathValue,
	}
	for key, value := range extra {
		updates[key] = value
	}
	return withEnv(os.Environ(), updates)
}

// installPnpm ensures Tavern's profile installer has the pinned package manager.
func installPnpm() error {
	versionCmd := exec.Command(pnpmBin(), "--version")
	versionOutput, versionErr := versionCmd.Output()
	if versionErr == nil && strings.TrimSpace(string(versionOutput)) == tavernPnpmVersion {
		return nil
	}
	if err := os.MkdirAll(globalPnpmDir, 0755); err != nil {
		return err
	}

	cfg := GetConfig()
	args := []string{"install", "pnpm@" + tavernPnpmVersion, "--save", "--no-audit", "--no-fund", "--registry=" + cfg.GetNpmRegistry()}
	cmd := exec.Command(npmBin(), args...)
	cmd.Dir = globalPnpmDir
	cmd.Stdout = NewLogWriterInfo()
	cmd.Stderr = NewLogWriterWarn()
	cmd.Env = tavernEnv(nil)
	LogInfo("正在初始化 Tavern pnpm 运行环境 (v%s)...", tavernPnpmVersion)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("安装 pnpm 失败: %w", err)
	}
	return nil
}

func installTavernDependencies() error {
	cmd := exec.Command(pnpmBin(), "--dir", tavernSourceDir, "install", "--frozen-lockfile")
	cmd.Stdout = NewLogWriterInfo()
	cmd.Stderr = NewLogWriterWarn()
	cmd.Env = tavernEnv(nil)
	LogInfo("正在安装 Tavern 程序依赖...")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("安装 Tavern 程序依赖失败: %w", err)
	}
	return nil
}

func runTavernCommand(args []string, extra map[string]string) error {
	script := filepath.Join(tavernSourceDir, "bin", "dsh-tavern.mjs")
	if _, err := os.Stat(script); err != nil {
		return fmt.Errorf("Tavern 源码未部署: %w", err)
	}
	cmdArgs := append([]string{script}, args...)
	cmd := exec.Command(nodeBin(), cmdArgs...)
	cmd.Dir = tavernSourceDir
	cmd.Stdout = NewLogWriterInfo()
	cmd.Stderr = NewLogWriterWarn()
	cmd.Env = tavernEnv(extra)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Tavern 命令失败 (%s): %w", strings.Join(args, " "), err)
	}
	return nil
}

func installTavern(forceRuntime bool) error {
	if err := installPnpm(); err != nil {
		return err
	}
	if err := installTavernDependencies(); err != nil {
		return err
	}
	extra := map[string]string{}
	if forceRuntime {
		extra["DSH_TAVERN_REINSTALL_RUNTIME"] = "1"
	}
	LogInfo("正在配置 DSH Tavern Profile...")
	return runTavernCommand([]string{"install", "--host", "cli"}, extra)
}

// CheckUpdateResult is intentionally commit based: Tavern is released from GitHub, not NPM.
type CheckUpdateResult struct {
	HasUpdate      bool   `json:"has_update"`
	CurrentVersion string `json:"current_version"`
	RemoteVersion  string `json:"remote_version"`
	Message        string `json:"message"`
}

type tavernCommit struct {
	SHA string `json:"sha"`
}

func shortCommit(value string) string {
	value = strings.TrimSpace(value)
	if len(value) > 8 {
		return value[:8]
	}
	return value
}

func installedTavernCommit() string {
	data, err := os.ReadFile(filepath.Join(tavernSourceDir, ".dsh-tavern-release.json"))
	if err != nil {
		return ""
	}
	var release struct {
		Commit string `json:"commit"`
	}
	if json.Unmarshal(data, &release) != nil {
		return ""
	}
	return strings.TrimSpace(release.Commit)
}

func fetchTavernCommit() (string, error) {
	transport := &http.Transport{Proxy: http.ProxyFromEnvironment}
	if proxy := strings.TrimSpace(GetConfig().NetworkProxy); proxy != "" {
		if proxyURL, err := url.Parse(proxy); err == nil {
			transport.Proxy = http.ProxyURL(proxyURL)
		}
	}
	client := &http.Client{Transport: transport, Timeout: 10 * time.Second}
	req, err := http.NewRequest(http.MethodGet, tavernCommitURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "DSH-Tavern-FNOS")
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub 返回 HTTP %d", resp.StatusCode)
	}
	var commit tavernCommit
	if err := json.NewDecoder(resp.Body).Decode(&commit); err != nil {
		return "", err
	}
	if len(commit.SHA) < 8 {
		return "", fmt.Errorf("GitHub 返回的提交号无效")
	}
	return commit.SHA, nil
}

func CheckUpdate() (*CheckUpdateResult, error) {
	current := installedTavernCommit()
	remote, err := fetchTavernCommit()
	if err != nil {
		return nil, fmt.Errorf("检查 Tavern 更新失败: %w", err)
	}
	currentVersion := shortCommit(current)
	remoteVersion := shortCommit(remote)
	hasUpdate := current == "" || !strings.EqualFold(current, remote)
	message := fmt.Sprintf("当前已是最新提交 [ %s ]", remoteVersion)
	if hasUpdate {
		message = fmt.Sprintf("发现 Tavern 新提交 [ %s → %s ]", currentVersionOrDash(currentVersion), remoteVersion)
	}
	LogInfo("[Tavern 更新] %s", message)
	return &CheckUpdateResult{
		HasUpdate:      hasUpdate,
		CurrentVersion: currentVersion,
		RemoteVersion:  remoteVersion,
		Message:        message,
	}, nil
}

func currentVersionOrDash(value string) string {
	if value == "" {
		return "-"
	}
	return value
}

func Upgrade() {
	state.SetStatus(StatusBuilding, "正在准备拉取 Tavern 更新...")
	go updateTavern(false)
}

func Rebuild() {
	state.SetStatus(StatusBuilding, "正在准备强制重建 Tavern 运行环境...")
	go updateTavern(true)
}

func safeRemoveAll(target string) error {
	if err := os.RemoveAll(target); err == nil {
		return nil
	}
	_ = filepath.Walk(target, func(path string, info os.FileInfo, err error) error {
		if err == nil {
			_ = os.Chmod(path, 0777)
		}
		return nil
	})
	return os.RemoveAll(target)
}

func RepairEnvironment(keepPlugins bool) {
	state.SetStatus(StatusBuilding, "正在准备恢复 Tavern 运行环境...")
	go func() {
		wasRunning := state.Status() == StatusRunning || GetLastRunState() == StatusRunning
		stopAndWait()
		if !keepPlugins {
			ResetAllProfilePatches()
		}
		_ = safeRemoveAll(filepath.Join(globalDshHome, "profiles", "tavern"))
		_ = safeRemoveAll(filepath.Join(tavernSourceDir, "node_modules"))
		if err := installTavern(false); err != nil {
			state.SetStatus(StatusStopped, "环境恢复失败: "+err.Error())
			return
		}
		refreshVersion()
		SetBuildTime(time.Now())
		state.SetStatus(StatusStopped, "")
		if wasRunning {
			restartService()
		}
	}()
}

func updateTavern(forceRebuild bool) {
	wasRunning := state.Status() == StatusRunning || GetLastRunState() == StatusRunning
	stopAndWait()

	var err error
	if forceRebuild {
		state.SetStatus(StatusBuilding, "正在强制重建 Tavern 依赖与独立 DSH...")
		_ = safeRemoveAll(filepath.Join(globalDshHome, "profiles", "tavern"))
		_ = safeRemoveAll(filepath.Join(tavernSourceDir, "node_modules"))
		_ = safeRemoveAll(filepath.Join(tavernSourceDir, "tavern-plugin", "node_modules"))
		err = installTavern(true)
	} else {
		state.SetStatus(StatusBuilding, "正在拉取 DSH Tavern 主分支并安装更新...")
		err = runTavernCommand([]string{"update", "--host", "cli"}, map[string]string{
			"DSH_TAVERN_NO_START":             "1",
			"DSH_TAVERN_FNOS_MANAGER_UPDATE": "1",
		})
	}
	if err != nil {
		LogWarning("Tavern 部署失败: %s", err)
		state.SetStatus(StatusStopped, "部署失败: "+err.Error())
		return
	}

	refreshVersion()
	SetBuildTime(time.Now())
	state.SetStatus(StatusStopped, "")
	LogInfo("Tavern 部署完成")
	if wasRunning {
		restartService()
	}
}

func extractTarGz(tarPath, dst string) error {
	if err := os.MkdirAll(dst, 0755); err != nil {
		return err
	}
	cmd := exec.Command("tar", "--no-same-owner", "-xzf", tarPath, "-C", dst)
	cmd.Stdout = NewLogWriterInfo()
	cmd.Stderr = NewLogWriterWarn()
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("解压归档失败: %w", err)
	}
	return nil
}

func refreshVersion() {
	version := readVersion()
	if version == "" {
		version = "-"
	}
	SetVersion(version)
}

func readVersion() string {
	data, err := os.ReadFile(filepath.Join(tavernSourceDir, "package.json"))
	if err != nil {
		return ""
	}
	var pkg struct {
		Version string `json:"version"`
	}
	if json.Unmarshal(data, &pkg) != nil {
		return ""
	}
	return strings.TrimSpace(pkg.Version)
}

func readAppDestVersion() string {
	data, err := os.ReadFile(filepath.Join(globalAppDest, ".version"))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}
