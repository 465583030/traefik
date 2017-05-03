package docker

import (
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/containous/traefik/types"
	"github.com/davecgh/go-spew/spew"
	dockerclient "github.com/docker/engine-api/client"
	docker "github.com/docker/engine-api/types"
	dockertypes "github.com/docker/engine-api/types"
	"github.com/docker/engine-api/types/swarm"
	"golang.org/x/net/context"
)

func TestDockerGetFrontendName(t *testing.T) {
	provider := &Docker{
		Domain: "docker.localhost",
	}

	containers := []struct {
		container docker.ContainerJSON
		expected  string
	}{
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "foo",
				},
				Config: &container.Config{},
			},
			expected: "Host-foo-docker-localhost",
		},
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "bar",
				},
				Config: &container.Config{
					Labels: map[string]string{
						"traefik.frontend.rule": "Headers:User-Agent,bat/0.1.0",
					},
				},
			},
			expected: "Headers-User-Agent-bat-0-1-0",
		},
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "test",
				},
				Config: &container.Config{
					Labels: map[string]string{
						"traefik.frontend.rule": "Host:foo.bar",
					},
				},
			},
			expected: "Host-foo-bar",
		},
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "test",
				},
				Config: &container.Config{
					Labels: map[string]string{
						"traefik.frontend.rule": "Path:/test",
					},
				},
			},
			expected: "Path-test",
		},
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "test",
				},
				Config: &container.Config{
					Labels: map[string]string{
						"traefik.frontend.rule": "PathPrefix:/test2",
					},
				},
			},
			expected: "PathPrefix-test2",
		},
	}

	for containerID, e := range containers {
		e := e
		t.Run(strconv.Itoa(containerID), func(t *testing.T) {
			t.Parallel()
			dockerData := parseContainer(e.container)
			actual := provider.getFrontendName(dockerData)
			if actual != e.expected {
				t.Errorf("expected %q, got %q", e.expected, actual)
			}
		})
	}
}

func TestDockerGetFrontendRule(t *testing.T) {
	provider := &Docker{
		Domain: "docker.localhost",
	}

	containers := []struct {
		container docker.ContainerJSON
		expected  string
	}{
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "foo",
				},
				Config: &container.Config{},
			},
			expected: "Host:foo.docker.localhost",
		},
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "bar",
				},
				Config: &container.Config{},
			},
			expected: "Host:bar.docker.localhost",
		},
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "test",
				},
				Config: &container.Config{
					Labels: map[string]string{
						"traefik.frontend.rule": "Host:foo.bar",
					},
				},
			},
			expected: "Host:foo.bar",
		},
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "test",
				},
				Config: &container.Config{
					Labels: map[string]string{
						"traefik.frontend.rule": "Path:/test",
					},
				},
			},
			expected: "Path:/test",
		},
	}

	for containerID, e := range containers {
		e := e
		t.Run(strconv.Itoa(containerID), func(t *testing.T) {
			t.Parallel()
			dockerData := parseContainer(e.container)
			actual := provider.getFrontendRule(dockerData)
			if actual != e.expected {
				t.Errorf("expected %q, got %q", e.expected, actual)
			}
		})
	}
}

func TestDockerGetBackend(t *testing.T) {
	provider := &Docker{}

	containers := []struct {
		container docker.ContainerJSON
		expected  string
	}{
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "foo",
				},
				Config: &container.Config{},
			},
			expected: "foo",
		},
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "bar",
				},
				Config: &container.Config{},
			},
			expected: "bar",
		},
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "test",
				},
				Config: &container.Config{
					Labels: map[string]string{
						"traefik.backend": "foobar",
					},
				},
			},
			expected: "foobar",
		},
	}

	for containerID, e := range containers {
		e := e
		t.Run(strconv.Itoa(containerID), func(t *testing.T) {
			t.Parallel()
			dockerData := parseContainer(e.container)
			actual := provider.getBackend(dockerData)
			if actual != e.expected {
				t.Errorf("expected %q, got %q", e.expected, actual)
			}
		})
	}
}

func TestDockerGetIPAddress(t *testing.T) { // TODO
	provider := &Docker{}

	containers := []struct {
		container docker.ContainerJSON
		expected  string
	}{
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "bar",
				},
				Config: &container.Config{},
				NetworkSettings: &docker.NetworkSettings{
					Networks: map[string]*network.EndpointSettings{
						"testnet": {
							IPAddress: "10.11.12.13",
						},
					},
				},
			},
			expected: "10.11.12.13",
		},
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "bar",
				},
				Config: &container.Config{
					Labels: map[string]string{
						"traefik.docker.network": "testnet",
					},
				},
				NetworkSettings: &docker.NetworkSettings{
					Networks: map[string]*network.EndpointSettings{
						"nottestnet": {
							IPAddress: "10.11.12.13",
						},
					},
				},
			},
			expected: "10.11.12.13",
		},
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "bar",
				},
				Config: &container.Config{
					Labels: map[string]string{
						"traefik.docker.network": "testnet2",
					},
				},
				NetworkSettings: &docker.NetworkSettings{
					Networks: map[string]*network.EndpointSettings{
						"testnet1": {
							IPAddress: "10.11.12.13",
						},
						"testnet2": {
							IPAddress: "10.11.12.14",
						},
					},
				},
			},
			expected: "10.11.12.14",
		},
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "bar",
					HostConfig: &container.HostConfig{
						NetworkMode: "host",
					},
				},
				Config: &container.Config{
					Labels: map[string]string{},
				},
				NetworkSettings: &docker.NetworkSettings{
					Networks: map[string]*network.EndpointSettings{
						"testnet1": {
							IPAddress: "10.11.12.13",
						},
						"testnet2": {
							IPAddress: "10.11.12.14",
						},
					},
				},
			},
			expected: "127.0.0.1",
		},
	}

	for containerID, e := range containers {
		e := e
		t.Run(strconv.Itoa(containerID), func(t *testing.T) {
			t.Parallel()
			dockerData := parseContainer(e.container)
			actual := provider.getIPAddress(dockerData)
			if actual != e.expected {
				t.Errorf("expected %q, got %q", e.expected, actual)
			}
		})
	}
}

func TestDockerGetPort(t *testing.T) {
	provider := &Docker{}

	containers := []struct {
		container docker.ContainerJSON
		expected  string
	}{
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "foo",
				},
				Config:          &container.Config{},
				NetworkSettings: &docker.NetworkSettings{},
			},
			expected: "",
		},
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "bar",
				},
				Config: &container.Config{},
				NetworkSettings: &docker.NetworkSettings{
					NetworkSettingsBase: docker.NetworkSettingsBase{
						Ports: nat.PortMap{
							"80/tcp": {},
						},
					},
				},
			},
			expected: "80",
		},
		// FIXME handle this better..
		// {
		// 	container: docker.ContainerJSON{
		// 		Name:   "bar",
		// 		Config: &container.Config{},
		// 		NetworkSettings: &docker.NetworkSettings{
		// 			Ports: map[docker.Port][]docker.PortBinding{
		// 				"80/tcp":  []docker.PortBinding{},
		// 				"443/tcp": []docker.PortBinding{},
		// 			},
		// 		},
		// 	},
		// 	expected: "80",
		// },
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "test",
				},
				Config: &container.Config{
					Labels: map[string]string{
						"traefik.port": "8080",
					},
				},
				NetworkSettings: &docker.NetworkSettings{},
			},
			expected: "8080",
		},
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "test",
				},
				Config: &container.Config{
					Labels: map[string]string{
						"traefik.port": "8080",
					},
				},
				NetworkSettings: &docker.NetworkSettings{
					NetworkSettingsBase: docker.NetworkSettingsBase{
						Ports: nat.PortMap{
							"80/tcp": {},
						},
					},
				},
			},
			expected: "8080",
		},
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "test-multi-ports",
				},
				Config: &container.Config{
					Labels: map[string]string{
						"traefik.port": "8080",
					},
				},
				NetworkSettings: &docker.NetworkSettings{
					NetworkSettingsBase: docker.NetworkSettingsBase{
						Ports: nat.PortMap{
							"8080/tcp": {},
							"80/tcp":   {},
						},
					},
				},
			},
			expected: "8080",
		},
	}

	for containerID, e := range containers {
		e := e
		t.Run(strconv.Itoa(containerID), func(t *testing.T) {
			t.Parallel()
			dockerData := parseContainer(e.container)
			actual := provider.getPort(dockerData)
			if actual != e.expected {
				t.Errorf("expected %q, got %q", e.expected, actual)
			}
		})
	}
}

func TestDockerGetWeight(t *testing.T) {
	provider := &Docker{}

	containers := []struct {
		container docker.ContainerJSON
		expected  string
	}{
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "foo",
				},
				Config: &container.Config{},
			},
			expected: "0",
		},
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "test",
				},
				Config: &container.Config{
					Labels: map[string]string{
						"traefik.weight": "10",
					},
				},
			},
			expected: "10",
		},
	}

	for containerID, e := range containers {
		e := e
		t.Run(strconv.Itoa(containerID), func(t *testing.T) {
			t.Parallel()
			dockerData := parseContainer(e.container)
			actual := provider.getWeight(dockerData)
			if actual != e.expected {
				t.Errorf("expected %q, got %q", e.expected, actual)
			}
		})
	}
}

func TestDockerGetDomain(t *testing.T) {
	provider := &Docker{
		Domain: "docker.localhost",
	}

	containers := []struct {
		container docker.ContainerJSON
		expected  string
	}{
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "foo",
				},
				Config: &container.Config{},
			},
			expected: "docker.localhost",
		},
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "test",
				},
				Config: &container.Config{
					Labels: map[string]string{
						"traefik.domain": "foo.bar",
					},
				},
			},
			expected: "foo.bar",
		},
	}

	for containerID, e := range containers {
		e := e
		t.Run(strconv.Itoa(containerID), func(t *testing.T) {
			t.Parallel()
			dockerData := parseContainer(e.container)
			actual := provider.getDomain(dockerData)
			if actual != e.expected {
				t.Errorf("expected %q, got %q", e.expected, actual)
			}
		})
	}
}

func TestDockerGetProtocol(t *testing.T) {
	provider := &Docker{}

	containers := []struct {
		container docker.ContainerJSON
		expected  string
	}{
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "foo",
				},
				Config: &container.Config{},
			},
			expected: "http",
		},
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "test",
				},
				Config: &container.Config{
					Labels: map[string]string{
						"traefik.protocol": "https",
					},
				},
			},
			expected: "https",
		},
	}

	for containerID, e := range containers {
		e := e
		t.Run(strconv.Itoa(containerID), func(t *testing.T) {
			t.Parallel()
			dockerData := parseContainer(e.container)
			actual := provider.getProtocol(dockerData)
			if actual != e.expected {
				t.Errorf("expected %q, got %q", e.expected, actual)
			}
		})
	}
}

func TestDockerGetPassHostHeader(t *testing.T) {
	provider := &Docker{}
	containers := []struct {
		container docker.ContainerJSON
		expected  string
	}{
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "foo",
				},
				Config: &container.Config{},
			},
			expected: "true",
		},
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "test",
				},
				Config: &container.Config{
					Labels: map[string]string{
						"traefik.frontend.passHostHeader": "false",
					},
				},
			},
			expected: "false",
		},
	}

	for containerID, e := range containers {
		e := e
		t.Run(strconv.Itoa(containerID), func(t *testing.T) {
			t.Parallel()
			dockerData := parseContainer(e.container)
			actual := provider.getPassHostHeader(dockerData)
			if actual != e.expected {
				t.Errorf("expected %q, got %q", e.expected, actual)
			}
		})
	}
}

func TestDockerGetLabel(t *testing.T) {
	containers := []struct {
		container docker.ContainerJSON
		expected  string
	}{
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "foo",
				},
				Config: &container.Config{},
			},
			expected: "Label not found:",
		},
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "foo",
				},
				Config: &container.Config{
					Labels: map[string]string{
						"foo": "bar",
					},
				},
			},
			expected: "",
		},
	}

	for containerID, e := range containers {
		e := e
		t.Run(strconv.Itoa(containerID), func(t *testing.T) {
			t.Parallel()
			dockerData := parseContainer(e.container)
			label, err := getLabel(dockerData, "foo")
			if e.expected != "" {
				if err == nil || !strings.Contains(err.Error(), e.expected) {
					t.Errorf("expected an error with %q, got %v", e.expected, err)
				}
			} else {
				if label != "bar" {
					t.Errorf("expected label 'bar', got %s", label)
				}
			}
		})
	}
}

func TestDockerGetLabels(t *testing.T) {
	containers := []struct {
		container      docker.ContainerJSON
		expectedLabels map[string]string
		expectedError  string
	}{
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "foo",
				},
				Config: &container.Config{},
			},
			expectedLabels: map[string]string{},
			expectedError:  "Label not found:",
		},
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "foo",
				},
				Config: &container.Config{
					Labels: map[string]string{
						"foo": "fooz",
					},
				},
			},
			expectedLabels: map[string]string{
				"foo": "fooz",
			},
			expectedError: "Label not found: bar",
		},
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "foo",
				},
				Config: &container.Config{
					Labels: map[string]string{
						"foo": "fooz",
						"bar": "barz",
					},
				},
			},
			expectedLabels: map[string]string{
				"foo": "fooz",
				"bar": "barz",
			},
			expectedError: "",
		},
	}

	for containerID, e := range containers {
		e := e
		t.Run(strconv.Itoa(containerID), func(t *testing.T) {
			t.Parallel()
			dockerData := parseContainer(e.container)
			labels, err := getLabels(dockerData, []string{"foo", "bar"})
			if !reflect.DeepEqual(labels, e.expectedLabels) {
				t.Errorf("expect %v, got %v", e.expectedLabels, labels)
			}
			if e.expectedError != "" {
				if err == nil || !strings.Contains(err.Error(), e.expectedError) {
					t.Errorf("expected an error with %q, got %v", e.expectedError, err)
				}
			}
		})
	}
}

func TestDockerTraefikFilter(t *testing.T) {
	containers := []struct {
		container docker.ContainerJSON
		expected  bool
		provider  *Docker
	}{
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "container",
				},
				Config:          &container.Config{},
				NetworkSettings: &docker.NetworkSettings{},
			},
			expected: false,
			provider: &Docker{
				Domain:           "test",
				ExposedByDefault: true,
			},
		},
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "container",
				},
				Config: &container.Config{
					Labels: map[string]string{
						"traefik.enable": "false",
					},
				},
				NetworkSettings: &docker.NetworkSettings{
					NetworkSettingsBase: docker.NetworkSettingsBase{
						Ports: nat.PortMap{
							"80/tcp": {},
						},
					},
				},
			},
			provider: &Docker{
				Domain:           "test",
				ExposedByDefault: true,
			},
			expected: false,
		},
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "container",
				},
				Config: &container.Config{
					Labels: map[string]string{
						"traefik.frontend.rule": "Host:foo.bar",
					},
				},
				NetworkSettings: &docker.NetworkSettings{
					NetworkSettingsBase: docker.NetworkSettingsBase{
						Ports: nat.PortMap{
							"80/tcp": {},
						},
					},
				},
			},
			provider: &Docker{
				Domain:           "test",
				ExposedByDefault: true,
			},
			expected: true,
		},
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "container-multi-ports",
				},
				Config: &container.Config{},
				NetworkSettings: &docker.NetworkSettings{
					NetworkSettingsBase: docker.NetworkSettingsBase{
						Ports: nat.PortMap{
							"80/tcp":  {},
							"443/tcp": {},
						},
					},
				},
			},
			provider: &Docker{
				Domain:           "test",
				ExposedByDefault: true,
			},
			expected: true,
		},
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "container",
				},
				Config: &container.Config{},
				NetworkSettings: &docker.NetworkSettings{
					NetworkSettingsBase: docker.NetworkSettingsBase{
						Ports: nat.PortMap{
							"80/tcp": {},
						},
					},
				},
			},
			provider: &Docker{
				Domain:           "test",
				ExposedByDefault: true,
			},
			expected: true,
		},
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "container",
				},
				Config: &container.Config{
					Labels: map[string]string{
						"traefik.port": "80",
					},
				},
				NetworkSettings: &docker.NetworkSettings{
					NetworkSettingsBase: docker.NetworkSettingsBase{
						Ports: nat.PortMap{
							"80/tcp":  {},
							"443/tcp": {},
						},
					},
				},
			},
			provider: &Docker{
				Domain:           "test",
				ExposedByDefault: true,
			},
			expected: true,
		},
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "container",
				},
				Config: &container.Config{
					Labels: map[string]string{
						"traefik.enable": "true",
					},
				},
				NetworkSettings: &docker.NetworkSettings{
					NetworkSettingsBase: docker.NetworkSettingsBase{
						Ports: nat.PortMap{
							"80/tcp": {},
						},
					},
				},
			},
			provider: &Docker{
				Domain:           "test",
				ExposedByDefault: true,
			},
			expected: true,
		},
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "container",
				},
				Config: &container.Config{
					Labels: map[string]string{
						"traefik.enable": "anything",
					},
				},
				NetworkSettings: &docker.NetworkSettings{
					NetworkSettingsBase: docker.NetworkSettingsBase{
						Ports: nat.PortMap{
							"80/tcp": {},
						},
					},
				},
			},
			provider: &Docker{
				Domain:           "test",
				ExposedByDefault: true,
			},
			expected: true,
		},
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "container",
				},
				Config: &container.Config{
					Labels: map[string]string{
						"traefik.frontend.rule": "Host:foo.bar",
					},
				},
				NetworkSettings: &docker.NetworkSettings{
					NetworkSettingsBase: docker.NetworkSettingsBase{
						Ports: nat.PortMap{
							"80/tcp": {},
						},
					},
				},
			},
			provider: &Docker{
				Domain:           "test",
				ExposedByDefault: true,
			},
			expected: true,
		},
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "container",
				},
				Config: &container.Config{},
				NetworkSettings: &docker.NetworkSettings{
					NetworkSettingsBase: docker.NetworkSettingsBase{
						Ports: nat.PortMap{
							"80/tcp": {},
						},
					},
				},
			},
			provider: &Docker{
				Domain:           "test",
				ExposedByDefault: false,
			},
			expected: false,
		},
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "container",
				},
				Config: &container.Config{
					Labels: map[string]string{
						"traefik.enable": "true",
					},
				},
				NetworkSettings: &docker.NetworkSettings{
					NetworkSettingsBase: docker.NetworkSettingsBase{
						Ports: nat.PortMap{
							"80/tcp": {},
						},
					},
				},
			},
			provider: &Docker{
				Domain:           "test",
				ExposedByDefault: false,
			},
			expected: true,
		},
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "container",
				},
				Config: &container.Config{
					Labels: map[string]string{
						"traefik.enable": "true",
					},
				},
				NetworkSettings: &docker.NetworkSettings{
					NetworkSettingsBase: docker.NetworkSettingsBase{
						Ports: nat.PortMap{
							"80/tcp": {},
						},
					},
				},
			},
			provider: &Docker{
				ExposedByDefault: false,
			},
			expected: false,
		},
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "container",
				},
				Config: &container.Config{
					Labels: map[string]string{
						"traefik.enable":        "true",
						"traefik.frontend.rule": "Host:i.love.this.host",
					},
				},
				NetworkSettings: &docker.NetworkSettings{
					NetworkSettingsBase: docker.NetworkSettingsBase{
						Ports: nat.PortMap{
							"80/tcp": {},
						},
					},
				},
			},
			provider: &Docker{
				ExposedByDefault: false,
			},
			expected: true,
		},
	}

	for containerID, e := range containers {
		e := e
		t.Run(strconv.Itoa(containerID), func(t *testing.T) {
			t.Parallel()
			dockerData := parseContainer(e.container)
			actual := e.provider.containerFilter(dockerData)
			if actual != e.expected {
				t.Errorf("expected %v for %+v, got %+v", e.expected, e, actual)
			}
		})
	}
}

func TestDockerLoadDockerConfig(t *testing.T) {
	cases := []struct {
		containers        []docker.ContainerJSON
		expectedFrontends map[string]*types.Frontend
		expectedBackends  map[string]*types.Backend
	}{
		{
			containers:        []docker.ContainerJSON{},
			expectedFrontends: map[string]*types.Frontend{},
			expectedBackends:  map[string]*types.Backend{},
		},
		{
			containers: []docker.ContainerJSON{
				{
					ContainerJSONBase: &docker.ContainerJSONBase{
						Name: "test",
					},
					Config: &container.Config{},
					NetworkSettings: &docker.NetworkSettings{
						NetworkSettingsBase: docker.NetworkSettingsBase{
							Ports: nat.PortMap{
								"80/tcp": {},
							},
						},
						Networks: map[string]*network.EndpointSettings{
							"bridge": {
								IPAddress: "127.0.0.1",
							},
						},
					},
				},
			},
			expectedFrontends: map[string]*types.Frontend{
				"frontend-Host-test-docker-localhost": {
					Backend:        "backend-test",
					PassHostHeader: true,
					EntryPoints:    []string{},
					Routes: map[string]types.Route{
						"route-frontend-Host-test-docker-localhost": {
							Rule: "Host:test.docker.localhost",
						},
					},
				},
			},
			expectedBackends: map[string]*types.Backend{
				"backend-test": {
					Servers: map[string]types.Server{
						"server-test": {
							URL:    "http://127.0.0.1:80",
							Weight: 0,
						},
					},
					CircuitBreaker: nil,
				},
			},
		},
		{
			containers: []docker.ContainerJSON{
				{
					ContainerJSONBase: &docker.ContainerJSONBase{
						Name: "test1",
					},
					Config: &container.Config{
						Labels: map[string]string{
							"traefik.backend":              "foobar",
							"traefik.frontend.entryPoints": "http,https",
						},
					},
					NetworkSettings: &docker.NetworkSettings{
						NetworkSettingsBase: docker.NetworkSettingsBase{
							Ports: nat.PortMap{
								"80/tcp": {},
							},
						},
						Networks: map[string]*network.EndpointSettings{
							"bridge": {
								IPAddress: "127.0.0.1",
							},
						},
					},
				},
				{
					ContainerJSONBase: &docker.ContainerJSONBase{
						Name: "test2",
					},
					Config: &container.Config{
						Labels: map[string]string{
							"traefik.backend": "foobar",
						},
					},
					NetworkSettings: &docker.NetworkSettings{
						NetworkSettingsBase: docker.NetworkSettingsBase{
							Ports: nat.PortMap{
								"80/tcp": {},
							},
						},
						Networks: map[string]*network.EndpointSettings{
							"bridge": {
								IPAddress: "127.0.0.1",
							},
						},
					},
				},
			},
			expectedFrontends: map[string]*types.Frontend{
				"frontend-Host-test1-docker-localhost": {
					Backend:        "backend-foobar",
					PassHostHeader: true,
					EntryPoints:    []string{"http", "https"},
					Routes: map[string]types.Route{
						"route-frontend-Host-test1-docker-localhost": {
							Rule: "Host:test1.docker.localhost",
						},
					},
				},
				"frontend-Host-test2-docker-localhost": {
					Backend:        "backend-foobar",
					PassHostHeader: true,
					EntryPoints:    []string{},
					Routes: map[string]types.Route{
						"route-frontend-Host-test2-docker-localhost": {
							Rule: "Host:test2.docker.localhost",
						},
					},
				},
			},
			expectedBackends: map[string]*types.Backend{
				"backend-foobar": {
					Servers: map[string]types.Server{
						"server-test1": {
							URL:    "http://127.0.0.1:80",
							Weight: 0,
						},
						"server-test2": {
							URL:    "http://127.0.0.1:80",
							Weight: 0,
						},
					},
					CircuitBreaker: nil,
				},
			},
		},
		{
			containers: []docker.ContainerJSON{
				{
					ContainerJSONBase: &docker.ContainerJSONBase{
						Name: "test1",
					},
					Config: &container.Config{
						Labels: map[string]string{
							"traefik.backend":                           "foobar",
							"traefik.frontend.entryPoints":              "http,https",
							"traefik.backend.maxconn.amount":            "1000",
							"traefik.backend.maxconn.extractorfunc":     "somethingelse",
							"traefik.backend.loadbalancer.method":       "drr",
							"traefik.backend.circuitbreaker.expression": "NetworkErrorRatio() > 0.5",
						},
					},
					NetworkSettings: &docker.NetworkSettings{
						NetworkSettingsBase: docker.NetworkSettingsBase{
							Ports: nat.PortMap{
								"80/tcp": {},
							},
						},
						Networks: map[string]*network.EndpointSettings{
							"bridge": {
								IPAddress: "127.0.0.1",
							},
						},
					},
				},
			},
			expectedFrontends: map[string]*types.Frontend{
				"frontend-Host-test1-docker-localhost": {
					Backend:        "backend-foobar",
					PassHostHeader: true,
					EntryPoints:    []string{"http", "https"},
					Routes: map[string]types.Route{
						"route-frontend-Host-test1-docker-localhost": {
							Rule: "Host:test1.docker.localhost",
						},
					},
				},
			},
			expectedBackends: map[string]*types.Backend{
				"backend-foobar": {
					Servers: map[string]types.Server{
						"server-test1": {
							URL:    "http://127.0.0.1:80",
							Weight: 0,
						},
					},
					CircuitBreaker: &types.CircuitBreaker{
						Expression: "NetworkErrorRatio() > 0.5",
					},
					LoadBalancer: &types.LoadBalancer{
						Method: "drr",
					},
					MaxConn: &types.MaxConn{
						Amount:        1000,
						ExtractorFunc: "somethingelse",
					},
				},
			},
		},
	}

	provider := &Docker{
		Domain:           "docker.localhost",
		ExposedByDefault: true,
	}

	for caseID, c := range cases {
		c := c
		t.Run(strconv.Itoa(caseID), func(t *testing.T) {
			t.Parallel()
			var dockerDataList []dockerData
			for _, container := range c.containers {
				dockerData := parseContainer(container)
				dockerDataList = append(dockerDataList, dockerData)
			}

			actualConfig := provider.loadDockerConfig(dockerDataList)
			// Compare backends
			if !reflect.DeepEqual(actualConfig.Backends, c.expectedBackends) {
				t.Errorf("expected %#v, got %#v", c.expectedBackends, actualConfig.Backends)
			}
			if !reflect.DeepEqual(actualConfig.Frontends, c.expectedFrontends) {
				t.Errorf("expected %#v, got %#v", c.expectedFrontends, actualConfig.Frontends)
			}
		})
	}
}

func TestSwarmGetFrontendName(t *testing.T) {
	services := []struct {
		service  swarm.Service
		expected string
		networks map[string]*docker.NetworkResource
	}{
		{
			service:  swarmService(serviceName("foo")),
			expected: "Host-foo-docker-localhost",
			networks: map[string]*docker.NetworkResource{},
		},
		{
			service: swarmService(serviceLabels(map[string]string{
				"traefik.frontend.rule": "Headers:User-Agent,bat/0.1.0",
			})),
			expected: "Headers-User-Agent-bat-0-1-0",
			networks: map[string]*docker.NetworkResource{},
		},
		{
			service: swarmService(serviceLabels(map[string]string{
				"traefik.frontend.rule": "Host:foo.bar",
			})),
			expected: "Host-foo-bar",
			networks: map[string]*docker.NetworkResource{},
		},
		{
			service: swarmService(serviceLabels(map[string]string{
				"traefik.frontend.rule": "Path:/test",
			})),
			expected: "Path-test",
			networks: map[string]*docker.NetworkResource{},
		},
		{
			service: swarmService(
				serviceName("test"),
				serviceLabels(map[string]string{
					"traefik.frontend.rule": "PathPrefix:/test2",
				}),
			),
			expected: "PathPrefix-test2",
			networks: map[string]*docker.NetworkResource{},
		},
	}

	for serviceID, e := range services {
		e := e
		t.Run(strconv.Itoa(serviceID), func(t *testing.T) {
			t.Parallel()
			dockerData := parseService(e.service, e.networks)
			provider := &Provider{
				Domain:    "docker.localhost",
				SwarmMode: true,
			}
			actual := provider.getFrontendName(dockerData)
			if actual != e.expected {
				t.Errorf("expected %q, got %q", e.expected, actual)
			}
		})
	}
}

func TestSwarmGetFrontendRule(t *testing.T) {
	services := []struct {
		service  swarm.Service
		expected string
		networks map[string]*docker.NetworkResource
	}{
		{
			service:  swarmService(serviceName("foo")),
			expected: "Host:foo.docker.localhost",
			networks: map[string]*docker.NetworkResource{},
		},
		{
			service:  swarmService(serviceName("bar")),
			expected: "Host:bar.docker.localhost",
			networks: map[string]*docker.NetworkResource{},
		},
		{
			service: swarmService(serviceLabels(map[string]string{
				"traefik.frontend.rule": "Host:foo.bar",
			})),
			expected: "Host:foo.bar",
			networks: map[string]*docker.NetworkResource{},
		},
		{
			service: swarmService(serviceLabels(map[string]string{
				"traefik.frontend.rule": "Path:/test",
			})),
			expected: "Path:/test",
			networks: map[string]*docker.NetworkResource{},
		},
	}

	for serviceID, e := range services {
		e := e
		t.Run(strconv.Itoa(serviceID), func(t *testing.T) {
			t.Parallel()
			dockerData := parseService(e.service, e.networks)
			provider := &Provider{
				Domain:    "docker.localhost",
				SwarmMode: true,
			}
			actual := provider.getFrontendRule(dockerData)
			if actual != e.expected {
				t.Errorf("expected %q, got %q", e.expected, actual)
			}
		})
	}
}

func TestSwarmGetBackend(t *testing.T) {
	services := []struct {
		service  swarm.Service
		expected string
		networks map[string]*docker.NetworkResource
	}{
		{
			service:  swarmService(serviceName("foo")),
			expected: "foo",
			networks: map[string]*docker.NetworkResource{},
		},
		{
			service:  swarmService(serviceName("bar")),
			expected: "bar",
			networks: map[string]*docker.NetworkResource{},
		},
		{
			service: swarmService(serviceLabels(map[string]string{
				"traefik.backend": "foobar",
			})),
			expected: "foobar",
			networks: map[string]*docker.NetworkResource{},
		},
	}

	for serviceID, e := range services {
		e := e
		t.Run(strconv.Itoa(serviceID), func(t *testing.T) {
			t.Parallel()
			dockerData := parseService(e.service, e.networks)
			provider := &Provider{
				SwarmMode: true,
			}
			actual := provider.getBackend(dockerData)
			if actual != e.expected {
				t.Errorf("expected %q, got %q", e.expected, actual)
			}
		})
	}
}

func TestSwarmGetIPAddress(t *testing.T) {
	services := []struct {
		service  swarm.Service
		expected string
		networks map[string]*docker.NetworkResource
	}{
		{
			service:  swarmService(withEndpointSpec(modeDNSSR)),
			expected: "",
			networks: map[string]*docker.NetworkResource{},
		},
		{
			service: swarmService(
				withEndpointSpec(modeVIP),
				withEndpoint(virtualIP("1", "10.11.12.13/24")),
			),
			expected: "10.11.12.13",
			networks: map[string]*docker.NetworkResource{
				"1": {
					Name: "foo",
				},
			},
		},
		{
			service: swarmService(
				serviceLabels(map[string]string{
					"traefik.docker.network": "barnet",
				}),
				withEndpointSpec(modeVIP),
				withEndpoint(
					virtualIP("1", "10.11.12.13/24"),
					virtualIP("2", "10.11.12.99/24"),
				),
			),
			expected: "10.11.12.99",
			networks: map[string]*docker.NetworkResource{
				"1": {
					Name: "foonet",
				},
				"2": {
					Name: "barnet",
				},
			},
		},
	}

	for serviceID, e := range services {
		e := e
		t.Run(strconv.Itoa(serviceID), func(t *testing.T) {
			t.Parallel()
			dockerData := parseService(e.service, e.networks)
			provider := &Provider{
				SwarmMode: true,
			}
			actual := provider.getIPAddress(dockerData)
			if actual != e.expected {
				t.Errorf("expected %q, got %q", e.expected, actual)
			}
		})
	}
}

func TestSwarmGetPort(t *testing.T) {
	services := []struct {
		service  swarm.Service
		expected string
		networks map[string]*docker.NetworkResource
	}{
		{
			service: swarmService(
				serviceLabels(map[string]string{
					"traefik.port": "8080",
				}),
				withEndpointSpec(modeDNSSR),
			),
			expected: "8080",
			networks: map[string]*docker.NetworkResource{},
		},
	}

	for serviceID, e := range services {
		e := e
		t.Run(strconv.Itoa(serviceID), func(t *testing.T) {
			t.Parallel()
			dockerData := parseService(e.service, e.networks)
			provider := &Provider{
				SwarmMode: true,
			}
			actual := provider.getPort(dockerData)
			if actual != e.expected {
				t.Errorf("expected %q, got %q", e.expected, actual)
			}
		})
	}
}

func TestSwarmGetWeight(t *testing.T) {
	services := []struct {
		service  swarm.Service
		expected string
		networks map[string]*docker.NetworkResource
	}{
		{
			service:  swarmService(),
			expected: "0",
			networks: map[string]*docker.NetworkResource{},
		},
		{
			service: swarmService(serviceLabels(map[string]string{
				"traefik.weight": "10",
			})),
			expected: "10",
			networks: map[string]*docker.NetworkResource{},
		},
	}

	for serviceID, e := range services {
		e := e
		t.Run(strconv.Itoa(serviceID), func(t *testing.T) {
			t.Parallel()
			dockerData := parseService(e.service, e.networks)
			provider := &Provider{
				SwarmMode: true,
			}
			actual := provider.getWeight(dockerData)
			if actual != e.expected {
				t.Errorf("expected %q, got %q", e.expected, actual)
			}
		})
	}
}

func TestSwarmGetDomain(t *testing.T) {
	services := []struct {
		service  swarm.Service
		expected string
		networks map[string]*docker.NetworkResource
	}{
		{
			service:  swarmService(serviceName("foo")),
			expected: "docker.localhost",
			networks: map[string]*docker.NetworkResource{},
		},
		{
			service: swarmService(serviceLabels(map[string]string{
				"traefik.domain": "foo.bar",
			})),
			expected: "foo.bar",
			networks: map[string]*docker.NetworkResource{},
		},
	}

	for serviceID, e := range services {
		e := e
		t.Run(strconv.Itoa(serviceID), func(t *testing.T) {
			t.Parallel()
			dockerData := parseService(e.service, e.networks)
			provider := &Provider{
				Domain:    "docker.localhost",
				SwarmMode: true,
			}
			actual := provider.getDomain(dockerData)
			if actual != e.expected {
				t.Errorf("expected %q, got %q", e.expected, actual)
			}
		})
	}
}

func TestSwarmGetProtocol(t *testing.T) {
	services := []struct {
		service  swarm.Service
		expected string
		networks map[string]*docker.NetworkResource
	}{
		{
			service:  swarmService(),
			expected: "http",
			networks: map[string]*docker.NetworkResource{},
		},
		{
			service: swarmService(serviceLabels(map[string]string{
				"traefik.protocol": "https",
			})),
			expected: "https",
			networks: map[string]*docker.NetworkResource{},
		},
	}

	for serviceID, e := range services {
		e := e
		t.Run(strconv.Itoa(serviceID), func(t *testing.T) {
			t.Parallel()
			dockerData := parseService(e.service, e.networks)
			provider := &Provider{
				SwarmMode: true,
			}
			actual := provider.getProtocol(dockerData)
			if actual != e.expected {
				t.Errorf("expected %q, got %q", e.expected, actual)
			}
		})
	}
}

func TestSwarmGetPassHostHeader(t *testing.T) {
	services := []struct {
		service  swarm.Service
		expected string
		networks map[string]*docker.NetworkResource
	}{
		{
			service:  swarmService(),
			expected: "true",
			networks: map[string]*docker.NetworkResource{},
		},
		{
			service: swarmService(serviceLabels(map[string]string{
				"traefik.frontend.passHostHeader": "false",
			})),
			expected: "false",
			networks: map[string]*docker.NetworkResource{},
		},
	}

	for serviceID, e := range services {
		e := e
		t.Run(strconv.Itoa(serviceID), func(t *testing.T) {
			t.Parallel()
			dockerData := parseService(e.service, e.networks)
			provider := &Provider{
				SwarmMode: true,
			}
			actual := provider.getPassHostHeader(dockerData)
			if actual != e.expected {
				t.Errorf("expected %q, got %q", e.expected, actual)
			}
		})
	}
}

func TestSwarmGetLabel(t *testing.T) {
	services := []struct {
		service  swarm.Service
		expected string
		networks map[string]*docker.NetworkResource
	}{
		{
			service:  swarmService(),
			expected: "Label not found:",
			networks: map[string]*docker.NetworkResource{},
		},
		{
			service: swarmService(serviceLabels(map[string]string{
				"foo": "bar",
			})),
			expected: "",
			networks: map[string]*docker.NetworkResource{},
		},
	}

	for serviceID, e := range services {
		e := e
		t.Run(strconv.Itoa(serviceID), func(t *testing.T) {
			t.Parallel()
			dockerData := parseService(e.service, e.networks)
			label, err := getLabel(dockerData, "foo")
			if e.expected != "" {
				if err == nil || !strings.Contains(err.Error(), e.expected) {
					t.Errorf("expected an error with %q, got %v", e.expected, err)
				}
			} else {
				if label != "bar" {
					t.Errorf("expected label 'bar', got %s", label)
				}
			}
		})
	}
}

func TestSwarmGetLabels(t *testing.T) {
	services := []struct {
		service        swarm.Service
		expectedLabels map[string]string
		expectedError  string
		networks       map[string]*docker.NetworkResource
	}{
		{
			service:        swarmService(),
			expectedLabels: map[string]string{},
			expectedError:  "Label not found:",
			networks:       map[string]*docker.NetworkResource{},
		},
		{
			service: swarmService(serviceLabels(map[string]string{
				"foo": "fooz",
			})),
			expectedLabels: map[string]string{
				"foo": "fooz",
			},
			expectedError: "Label not found: bar",
			networks:      map[string]*docker.NetworkResource{},
		},
		{
			service: swarmService(serviceLabels(map[string]string{
				"foo": "fooz",
				"bar": "barz",
			})),
			expectedLabels: map[string]string{
				"foo": "fooz",
				"bar": "barz",
			},
			expectedError: "",
			networks:      map[string]*docker.NetworkResource{},
		},
	}

	for serviceID, e := range services {
		e := e
		t.Run(strconv.Itoa(serviceID), func(t *testing.T) {
			t.Parallel()
			dockerData := parseService(e.service, e.networks)
			labels, err := getLabels(dockerData, []string{"foo", "bar"})
			if !reflect.DeepEqual(labels, e.expectedLabels) {
				t.Errorf("expect %v, got %v", e.expectedLabels, labels)
			}
			if e.expectedError != "" {
				if err == nil || !strings.Contains(err.Error(), e.expectedError) {
					t.Errorf("expected an error with %q, got %v", e.expectedError, err)
				}
			}
		})
	}
}

func TestSwarmTraefikFilter(t *testing.T) {
	services := []struct {
		service  swarm.Service
		expected bool
		networks map[string]*docker.NetworkResource
		provider *Docker
	}{
		{
			service:  swarmService(),
			expected: false,
			networks: map[string]*docker.NetworkResource{},
			provider: &Docker{
				SwarmMode:        true,
				Domain:           "test",
				ExposedByDefault: true,
			},
		},
		{
			service: swarmService(serviceLabels(map[string]string{
				"traefik.enable": "false",
				"traefik.port":   "80",
			})),
			expected: false,
			networks: map[string]*docker.NetworkResource{},
			provider: &Docker{
				SwarmMode:        true,
				Domain:           "test",
				ExposedByDefault: true,
			},
		},
		{
			service: swarmService(serviceLabels(map[string]string{
				"traefik.frontend.rule": "Host:foo.bar",
				"traefik.port":          "80",
			})),
			expected: true,
			networks: map[string]*docker.NetworkResource{},
			provider: &Docker{
				SwarmMode:        true,
				Domain:           "test",
				ExposedByDefault: true,
			},
		},
		{
			service: swarmService(serviceLabels(map[string]string{
				"traefik.port": "80",
			})),
			expected: true,
			networks: map[string]*docker.NetworkResource{},
			provider: &Docker{
				SwarmMode:        true,
				Domain:           "test",
				ExposedByDefault: true,
			},
		},
		{
			service: swarmService(serviceLabels(map[string]string{
				"traefik.enable": "true",
				"traefik.port":   "80",
			})),
			expected: true,
			networks: map[string]*docker.NetworkResource{},
			provider: &Docker{
				SwarmMode:        true,
				Domain:           "test",
				ExposedByDefault: true,
			},
		},
		{
			service: swarmService(serviceLabels(map[string]string{
				"traefik.enable": "anything",
				"traefik.port":   "80",
			})),
			expected: true,
			networks: map[string]*docker.NetworkResource{},
			provider: &Docker{
				SwarmMode:        true,
				Domain:           "test",
				ExposedByDefault: true,
			},
		},
		{
			service: swarmService(serviceLabels(map[string]string{
				"traefik.frontend.rule": "Host:foo.bar",
				"traefik.port":          "80",
			})),
			expected: true,
			networks: map[string]*docker.NetworkResource{},
			provider: &Docker{
				SwarmMode:        true,
				Domain:           "test",
				ExposedByDefault: true,
			},
		},
		{
			service: swarmService(serviceLabels(map[string]string{
				"traefik.port": "80",
			})),
			expected: false,
			networks: map[string]*docker.NetworkResource{},
			provider: &Docker{
				SwarmMode:        true,
				Domain:           "test",
				ExposedByDefault: false,
			},
		},
		{
			service: swarmService(serviceLabels(map[string]string{
				"traefik.enable": "true",
				"traefik.port":   "80",
			})),
			expected: true,
			networks: map[string]*docker.NetworkResource{},
			provider: &Docker{
				SwarmMode:        true,
				Domain:           "test",
				ExposedByDefault: false,
			},
		},
	}

	for serviceID, e := range services {
		e := e
		t.Run(strconv.Itoa(serviceID), func(t *testing.T) {
			t.Parallel()
			dockerData := parseService(e.service, e.networks)
			provider.ExposedByDefault = e.exposedByDefault
			actual := provider.containerFilter(dockerData)
			if actual != e.expected {
				t.Errorf("expected %v for %+v, got %+v", e.expected, e, actual)
			}
		})
	}
}

func TestSwarmLoadDockerConfig(t *testing.T) {
	cases := []struct {
		services          []swarm.Service
		expectedFrontends map[string]*types.Frontend
		expectedBackends  map[string]*types.Backend
		networks          map[string]*docker.NetworkResource
	}{
		{
			services:          []swarm.Service{},
			expectedFrontends: map[string]*types.Frontend{},
			expectedBackends:  map[string]*types.Backend{},
			networks:          map[string]*docker.NetworkResource{},
		},
		{
			services: []swarm.Service{
				swarmService(
					serviceName("test"),
					serviceLabels(map[string]string{
						"traefik.port": "80",
					}),
					withEndpointSpec(modeVIP),
					withEndpoint(virtualIP("1", "127.0.0.1/24")),
				),
			},
			expectedFrontends: map[string]*types.Frontend{
				"frontend-Host-test-docker-localhost": {
					Backend:        "backend-test",
					PassHostHeader: true,
					EntryPoints:    []string{},
					BasicAuth:      []string{},
					Routes: map[string]types.Route{
						"route-frontend-Host-test-docker-localhost": {
							Rule: "Host:test.docker.localhost",
						},
					},
				},
			},
			expectedBackends: map[string]*types.Backend{
				"backend-test": {
					Servers: map[string]types.Server{
						"server-test": {
							URL:    "http://127.0.0.1:80",
							Weight: 0,
						},
					},
					CircuitBreaker: nil,
					LoadBalancer:   nil,
				},
			},
			networks: map[string]*docker.NetworkResource{
				"1": {
					Name: "foo",
				},
			},
		},
		{
			services: []swarm.Service{
				swarmService(
					serviceName("test1"),
					serviceLabels(map[string]string{
						"traefik.port":                 "80",
						"traefik.backend":              "foobar",
						"traefik.frontend.entryPoints": "http,https",
						"traefik.frontend.auth.basic":  "test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/,test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0",
					}),
					withEndpointSpec(modeVIP),
					withEndpoint(virtualIP("1", "127.0.0.1/24")),
				),
				swarmService(
					serviceName("test2"),
					serviceLabels(map[string]string{
						"traefik.port":    "80",
						"traefik.backend": "foobar",
					}),
					withEndpointSpec(modeVIP),
					withEndpoint(virtualIP("1", "127.0.0.1/24")),
				),
			},
			expectedFrontends: map[string]*types.Frontend{
				"frontend-Host-test1-docker-localhost": {
					Backend:        "backend-foobar",
					PassHostHeader: true,
					EntryPoints:    []string{"http", "https"},
					BasicAuth:      []string{"test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/", "test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0"},
					Routes: map[string]types.Route{
						"route-frontend-Host-test1-docker-localhost": {
							Rule: "Host:test1.docker.localhost",
						},
					},
				},
				"frontend-Host-test2-docker-localhost": {
					Backend:        "backend-foobar",
					PassHostHeader: true,
					EntryPoints:    []string{},
					BasicAuth:      []string{},
					Routes: map[string]types.Route{
						"route-frontend-Host-test2-docker-localhost": {
							Rule: "Host:test2.docker.localhost",
						},
					},
				},
			},
			expectedBackends: map[string]*types.Backend{
				"backend-foobar": {
					Servers: map[string]types.Server{
						"server-test1": {
							URL:    "http://127.0.0.1:80",
							Weight: 0,
						},
						"server-test2": {
							URL:    "http://127.0.0.1:80",
							Weight: 0,
						},
					},
					CircuitBreaker: nil,
					LoadBalancer:   nil,
				},
			},
			networks: map[string]*docker.NetworkResource{
				"1": {
					Name: "foo",
				},
			},
		},
	}

	for caseID, c := range cases {
		c := c
		t.Run(strconv.Itoa(caseID), func(t *testing.T) {
			t.Parallel()
			var dockerDataList []dockerData
			for _, service := range c.services {
				dockerData := parseService(service, c.networks)
				dockerDataList = append(dockerDataList, dockerData)
			}

			provider := &Provider{
				Domain:           "docker.localhost",
				ExposedByDefault: true,
				SwarmMode:        true,
			}
			actualConfig := provider.loadDockerConfig(dockerDataList)
			// Compare backends
			if !reflect.DeepEqual(actualConfig.Backends, c.expectedBackends) {
				t.Errorf("expected %#v, got %#v", c.expectedBackends, actualConfig.Backends)
			}
			if !reflect.DeepEqual(actualConfig.Frontends, c.expectedFrontends) {
				t.Errorf("expected %#v, got %#v", c.expectedFrontends, actualConfig.Frontends)
			}
		})
	}
}

func TestSwarmTaskParsing(t *testing.T) {
	cases := []struct {
		service       swarm.Service
		tasks         []swarm.Task
		isGlobalSVC   bool
		expectedNames map[string]string
		networks      map[string]*docker.NetworkResource
	}{
		{
			service: swarmService(serviceName("container")),
			tasks: []swarm.Task{
				swarmTask("id1", taskSlot(1)),
				swarmTask("id2", taskSlot(2)),
				swarmTask("id3", taskSlot(3)),
			},
			isGlobalSVC: false,
			expectedNames: map[string]string{
				"id1": "container.1",
				"id2": "container.2",
				"id3": "container.3",
			},
			networks: map[string]*docker.NetworkResource{
				"1": {
					Name: "foo",
				},
			},
		},
		{
			service: swarmService(serviceName("container")),
			tasks: []swarm.Task{
				swarmTask("id1"),
				swarmTask("id2"),
				swarmTask("id3"),
			},
			isGlobalSVC: true,
			expectedNames: map[string]string{
				"id1": "container.id1",
				"id2": "container.id2",
				"id3": "container.id3",
			},
			networks: map[string]*docker.NetworkResource{
				"1": {
					Name: "foo",
				},
			},
		},
	}

	for caseID, e := range cases {
		e := e
		t.Run(strconv.Itoa(caseID), func(t *testing.T) {
			t.Parallel()
			dockerData := parseService(e.service, e.networks)

			for _, task := range e.tasks {
				taskDockerData := parseTasks(task, dockerData, map[string]*docker.NetworkResource{}, e.isGlobalSVC)
				if !reflect.DeepEqual(taskDockerData.Name, e.expectedNames[task.ID]) {
					t.Errorf("expect %v, got %v", e.expectedNames[task.ID], taskDockerData.Name)
				}
			}
		})
	}
}

type fakeTasksClient struct {
	dockerclient.APIClient
	tasks []swarm.Task
	err   error
}

func (c *fakeTasksClient) TaskList(ctx context.Context, options dockertypes.TaskListOptions) ([]swarm.Task, error) {
	return c.tasks, c.err
}

func TestListTasks(t *testing.T) {
	cases := []struct {
		service       swarm.Service
		tasks         []swarm.Task
		isGlobalSVC   bool
		expectedTasks []string
		networks      map[string]*docker.NetworkResource
	}{
		{
			service: swarmService(serviceName("container")),
			tasks: []swarm.Task{
				swarmTask("id1", taskSlot(1), taskStatus(taskState(swarm.TaskStateRunning))),
				swarmTask("id2", taskSlot(2), taskStatus(taskState(swarm.TaskStatePending))),
				swarmTask("id3", taskSlot(3)),
				swarmTask("id4", taskSlot(4), taskStatus(taskState(swarm.TaskStateRunning))),
				swarmTask("id5", taskSlot(5), taskStatus(taskState(swarm.TaskStateFailed))),
			},
			isGlobalSVC: false,
			expectedTasks: []string{
				"container.1",
				"container.4",
			},
			networks: map[string]*docker.NetworkResource{
				"1": {
					Name: "foo",
				},
			},
		},
	}

	for caseID, e := range cases {
		e := e
		t.Run(strconv.Itoa(caseID), func(t *testing.T) {
			t.Parallel()
			dockerData := parseService(e.service, e.networks)
			dockerClient := &fakeTasksClient{tasks: e.tasks}
			taskDockerData, _ := listTasks(context.Background(), dockerClient, e.service.ID, dockerData, map[string]*docker.NetworkResource{}, e.isGlobalSVC)

			if len(e.expectedTasks) != len(taskDockerData) {
				t.Errorf("expected tasks %v, got %v", spew.Sdump(e.expectedTasks), spew.Sdump(taskDockerData))
			}

			for i, taskID := range e.expectedTasks {
				if taskDockerData[i].Name != taskID {
					t.Errorf("expect task id %v, got %v", taskID, taskDockerData[i].Name)
				}
			}
		})
	}
}
