package mirror

// Pre-configured Chinese mirror sources

// Pip mirrors
var PipMirrors = []MirrorConfig{
	{Name: "阿里云", URL: "https://mirrors.aliyun.com/pypi/simple/"},
	{Name: "腾讯云", URL: "https://mirrors.cloud.tencent.com/pypi/simple/"},
	{Name: "华为云", URL: "https://mirrors.huaweicloud.com/repository/pypi/simple/"},
	{Name: "清华大学", URL: "https://pypi.tuna.tsinghua.edu.cn/simple"},
}

// Maven mirrors
var MavenMirrors = []MirrorConfig{
	{Name: "阿里云", URL: "https://maven.aliyun.com/repository/public"},
	{Name: "腾讯云", URL: "https://mirrors.cloud.tencent.com/nexus/repository/maven-public/"},
	{Name: "华为云", URL: "https://mirrors.huaweicloud.com/repository/maven/"},
}

// npm mirrors
var NpmMirrors = []MirrorConfig{
	{Name: "阿里云", URL: "https://registry.npmmirror.com"},
	{Name: "腾讯云", URL: "https://mirrors.cloud.tencent.com/npm/"},
	{Name: "华为云", URL: "https://mirrors.huaweicloud.com/repository/npm/"},
	{Name: "淘宝镜像", URL: "https://registry.npmmirror.com"},
}

// GetMirrorNames returns the names of available mirrors for a package manager
func GetMirrorNames(managerName string) []string {
	switch managerName {
	case "pip":
		return extractNames(PipMirrors)
	case "maven":
		return extractNames(MavenMirrors)
	case "npm":
		return extractNames(NpmMirrors)
	default:
		return []string{}
	}
}

// GetMirrorByName returns the mirror config by name for a package manager
func GetMirrorByName(managerName, mirrorName string) (MirrorConfig, bool) {
	var mirrors []MirrorConfig
	switch managerName {
	case "pip":
		mirrors = PipMirrors
	case "maven":
		mirrors = MavenMirrors
	case "npm":
		mirrors = NpmMirrors
	default:
		return MirrorConfig{}, false
	}

	for _, m := range mirrors {
		if m.Name == mirrorName {
			return m, true
		}
	}
	return MirrorConfig{}, false
}

// extractNames extracts names from mirror configs
func extractNames(mirrors []MirrorConfig) []string {
	names := make([]string, len(mirrors))
	for i, m := range mirrors {
		names[i] = m.Name
	}
	return names
}
