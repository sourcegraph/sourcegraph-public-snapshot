package parser

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
)

var (
	ErrImageMissing   = errors.New("Yaml must specify an image for every step")
	ErrImageWhitelist = errors.New("Yaml must specify am image from the white-list")
)

var (
	DefaultCloner = "plugins/drone-git"   // Default clone plugin.
	DefaultCacher = "plugins/drone-cache" // Default cache plugin.
	DefaultMatch  = "plugins/*"           // Default plugin whitelist.
)

// RuleFunc defines a function used to validate or modify the yaml during
// the parsing process.
type RuleFunc func(Node) error

// ImageName expands to a fully qualified image name. If no image name is found,
// a default is used when possible, else ErrImageMissing is returned.
func ImageName(n Node) error {
	d, ok := n.(*DockerNode)
	if !ok {
		return nil
	}

	switch d.NodeType {
	case NodeCompose:
		break
	case NodeBuild:
		// some projects may use drone solely for deployments
		// and therefore have no build section. Only enforce
		// the image if commands exist.
		if len(d.Image) == 0 && len(d.Commands) == 0 {
			return nil
		}
		break
	case NodeClone:
		d.Image = expandImageDefault(d.Image, DefaultCloner)
	case NodeCache:
		d.Image = expandImageDefault(d.Image, DefaultCacher)
	default:
		d.Image = expandImage(d.Image)
	}

	if len(d.Image) == 0 {
		return ErrImageMissing
	}
	d.Image = expandImageTag(d.Image)
	return nil
}

// ImageMatch checks the image name against a whitelist.
func ImageMatch(n Node, patterns []string) error {
	d, ok := n.(*DockerNode)
	if !ok {
		return nil
	}
	switch d.NodeType {
	case NodeBuild, NodeCompose:
		return nil
	}
	if len(patterns) == 0 || patterns[0] == "" {
		patterns = []string{DefaultMatch}
	}
	match := false
	for _, pattern := range patterns {
		if pattern == d.Image {
			match = true
			break
		}
		ok, err := filepath.Match(pattern, d.Image)
		if ok && err == nil {
			match = true
			break
		}
	}
	if !match {
		return fmt.Errorf("Plugin %s is not in the whitelist", d.Image)
	}
	return nil
}

func ImageMatchFunc(patterns []string) RuleFunc {
	return func(n Node) error {
		return ImageMatch(n, patterns)
	}
}

func ImagePull(n Node, pull bool) error {
	d, ok := n.(*DockerNode)
	if !ok {
		return nil
	}
	switch d.NodeType {
	case NodeBuild, NodeCompose:
		return nil
	}
	d.Pull = pull
	return nil
}

func ImagePullFunc(pull bool) RuleFunc {
	return func(n Node) error {
		return ImagePull(n, pull)
	}
}

// Sanitize sanitizes a Docker Node by removing any potentially
// harmful configuration options.
func Sanitize(n Node) error {
	d, ok := n.(*DockerNode)
	if !ok {
		return nil
	}
	d.Privileged = false
	d.Volumes = nil
	d.Net = ""
	d.Entrypoint = []string{}
	return nil
}

func SanitizeFunc(trusted bool) RuleFunc {
	return func(n Node) error {
		if !trusted {
			return Sanitize(n)
		}
		return nil
	}
}

// Escalate escalates a Docker Node to run in privileged mode if
// the plugin is whitelisted.
func Escalate(n Node) error {
	d, ok := n.(*DockerNode)
	if !ok {
		return nil
	}
	image := strings.Split(d.Image, ":")
	if d.NodeType == NodePublish && (image[0] == "plugins/drone-docker" || image[0] == "plugins/drone-gcr") {

		d.Privileged = true
		d.Volumes = nil
		d.Net = ""
		d.Entrypoint = []string{}
		//d.Volumes = []string{"/lib/modules:/lib/modules"}
	}
	return nil
}

func DefaultNotifyFilter(n Node) error {
	f, ok := n.(*FilterNode)
	if !ok || f.Node == nil {
		return nil
	}

	d, ok := f.Node.(*DockerNode)
	if !ok {
		return nil
	}
	if d.NodeType != NodeNotify {
		return nil
	}
	empty := len(f.Success) == 0 &&
		len(f.Failure) == 0 &&
		len(f.Change) == 0

	if !empty {
		if len(f.Success) == 0 {
			f.Success = "false"
		}
		if len(f.Failure) == 0 {
			f.Failure = "false"
		}
		if len(f.Change) == 0 {
			f.Change = "false"
		}
	}
	return nil
}

// HttpProxy injects the HTTP_PROXY and HTTPS_PROXY environment
// variables into the container.
func HttpProxy(n Node) error {
	d, ok := n.(*DockerNode)
	if !ok {
		return nil
	}
	for _, env := range os.Environ() {
		envup := strings.ToUpper(env)
		if strings.HasPrefix(envup, "HTTP_PROXY") ||
			strings.HasPrefix(envup, "HTTPS_PROXY") ||
			strings.HasPrefix(envup, "NO_PROXY") {
			d.Environment = append([]string{env}, d.Environment...)
		}
	}
	return nil
}

// Cache transforms the Docker Node to mount a volume to the host
// machines local cache.
func Cache(n Node, dir string) error {
	d, ok := n.(*DockerNode)
	if !ok {
		return nil
	}
	if d.NodeType == NodeCache {
		dir = fmt.Sprintf("/var/lib/drone/cache/%s:/cache", dir)
		d.Volumes = []string{dir}
	}
	return nil
}

func CacheFunc(dir string) RuleFunc {
	return func(n Node) error {
		return Cache(n, dir)
	}
}

// Debug transforms plugin Nodes to set the DEBUG environment
// variable when starting plugins.
func Debug(n Node, debug bool) error {
	d, ok := n.(*DockerNode)
	if !ok {
		return nil
	}
	switch d.NodeType {
	case NodeCache, NodeClone, NodeDeploy, NodeNotify, NodePublish:
		d.Environment = append(d.Environment, "DEBUG=true")
	}
	return nil
}

func DebugFunc(debug bool) RuleFunc {
	return func(n Node) error {
		return Debug(n, debug)
	}
}

func Mount(n Node, from, to string) error {
	d, ok := n.(*DockerNode)
	if !ok {
		return nil
	}
	dir := fmt.Sprintf("%s:%s", from, to)
	d.Volumes = append(d.Volumes, dir)
	return nil
}

func MountFunc(from, to string) RuleFunc {
	return func(n Node) error {
		return Mount(n, from, to)
	}
}

// expandImage expands an alias plugin name to use a
// fully qualified image name.
func expandImage(image string) string {
	if !strings.Contains(image, "/") {
		image = path.Join("plugins", "drone-"+image)
	}
	return strings.Replace(image, "_", "-", -1)
}

// expandImageDefault returns the default image if none
// is specified in the Yaml. If an image is specified,
// it expands the alias.
func expandImageDefault(image, defaultImage string) string {
	if len(image) == 0 {
		return defaultImage
	}
	return expandImage(image)
}

// expandImageTag is a helper function that automatically
// expands the image to include the :latest tag if not present
func expandImageTag(image string) string {
	if strings.Contains(image, "@") {
		return image
	}
	n := strings.LastIndex(image, ":")
	if n < 0 {
		return image + ":latest"
	}
	if tag := image[n+1:]; strings.Contains(tag, "/") {
		return image + ":latest"
	}
	return image
}
