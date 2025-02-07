package values

import (
	"github.com/loft-sh/vcluster/cmd/vclusterctl/cmd/app/create"
	"github.com/loft-sh/vcluster/cmd/vclusterctl/log"
	"k8s.io/client-go/kubernetes"
	"strings"
)

var K0SVersionMap = map[string]string{
	"1.22": "k0sproject/k0s:v1.22.4-k0s.0",
}

func getDefaultK0SReleaseValues(client kubernetes.Interface, createOptions *create.CreateOptions, log log.Logger) (string, error) {
	serverVersionString, serverMinorInt, err := getKubernetesVersion(client, createOptions)
	if err != nil {
		return "", err
	}

	image, ok := K0SVersionMap[serverVersionString]
	if !ok {
		if serverMinorInt > 22 {
			log.Infof("officially unsupported host server version %s, will fallback to virtual cluster version v1.22", serverVersionString)
			image = K0SVersionMap["1.22"]
		} else {
			log.Infof("officially unsupported host server version %s, will fallback to virtual cluster version v1.22", serverVersionString)
			image = K0SVersionMap["1.22"]
		}
	}

	// build values
	values := `vcluster:
  image: ##IMAGE##
`
	values = strings.ReplaceAll(values, "##IMAGE##", image)
	return addCommonReleaseValues(values, createOptions)
}
